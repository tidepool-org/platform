package events

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	logjson "github.com/tidepool-org/platform/log/json"
	lognull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/push"
)

type Consumer struct {
	Alerts             AlertsClient
	Data               alerts.DataRepository
	DeviceTokens       auth.DeviceTokensClient
	Evaluator          AlertsEvaluator
	Permissions        permission.Client
	Pusher             Pusher
	LastCommunications LastCommunicationsRecorder
	TokensProvider     auth.ServerSessionTokenProvider

	Logger log.Logger
}

// DosingDecision removes a stutter to improve readability.
type DosingDecision = dosingdecision.DosingDecision

// Glucose removes a stutter to improve readability.
type Glucose = glucose.Glucose

func (c *Consumer) Consume(ctx context.Context,
	session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) (err error) {

	if msg == nil {
		c.logger(ctx).Info("UNEXPECTED: nil message; ignoring")
		return nil
	}

	ctx = auth.NewContextWithServerSessionTokenProvider(ctx, c.TokensProvider)

	switch {
	case strings.Contains(msg.Topic, ".data.alerts"):
		return c.consumeAlertsConfigs(ctx, session, msg)
	case strings.Contains(msg.Topic, ".data.deviceData.alerts"):
		return c.consumeDeviceData(ctx, session, msg)
	default:
		c.logger(ctx).WithField("topic", msg.Topic).
			Infof("UNEXPECTED: topic; ignoring")
	}

	return nil
}

func (c *Consumer) consumeAlertsConfigs(ctx context.Context,
	session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {

	cfg := &alerts.Config{}
	updatedFields, err := unmarshalMessageValue(msg.Value, cfg)
	if err != nil {
		return err
	}
	lgr := c.logger(ctx)
	if isActivityAndActivityOnly(updatedFields) {
		lgr.WithField("updatedFields", updatedFields).
			Debug("alerts config is an activity update, will skip")
		lgr.WithField("message", msg).Debug("marked")
		session.MarkMessage(msg, "")
		return nil
	}

	lgr.WithField("cfg", cfg).Info("consuming an alerts config message")

	ctxLog := cfg.LoggerWithFields(c.logger(ctx))
	ctx = log.NewContextWithLogger(ctx, ctxLog)

	notes, err := c.Evaluator.EvaluateData(ctx, cfg.FollowedUserID, cfg.UploadID)
	if err != nil {
		format := "Unable to evalaute alerts configs triggered event for user %s"
		return errors.Wrapf(err, format, cfg.UserID)
	}
	ctxLog.WithField("notes", notes).Debug("notes generated from alerts config")

	c.pushNotifications(ctx, notes)

	session.MarkMessage(msg, "")
	lgr.WithField("message", msg).Debug("marked")
	return nil
}

func isActivityAndActivityOnly(updatedFields []string) bool {
	hasActivity := false
	for _, field := range updatedFields {
		if field == "activity" {
			hasActivity = true
		} else {
			return false
		}
	}
	return hasActivity
}

func (c *Consumer) consumeDeviceData(ctx context.Context,
	session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {

	lgr := c.logger(ctx)
	lgr.Debug("consuming device data message")

	// The actual type should be either a glucose.Glucose or a
	// dosingdecision.DosingDecision, but they both use types.Base, and that's where the
	// only fields we need are defined.
	datum := &types.Base{}
	if _, err := unmarshalMessageValue(msg.Value, datum); err != nil {
		return err
	}
	if datum.UserID == nil {
		return errors.New("Unable to retrieve alerts configs: userID is nil")
	}
	if datum.UploadID == nil {
		return errors.New("Unable to retrieve alerts configs: DataSetID is nil")
	}
	ctx = log.NewContextWithLogger(ctx, lgr.WithField("followedUserID", *datum.UserID))
	lastComm := alerts.LastCommunication{
		UserID:                 *datum.UserID,
		LastReceivedDeviceData: time.Now(),
		DataSetID:              *datum.UploadID,
	}
	err := c.LastCommunications.RecordReceivedDeviceData(ctx, lastComm)
	if err != nil {
		lgr.WithError(err).Info("Unable to record device data received")
	}
	notes, err := c.Evaluator.EvaluateData(ctx, *datum.UserID, *datum.UploadID)
	if err != nil {
		format := "Unable to evalaute device data triggered event for user %s"
		return errors.Wrapf(err, format, *datum.UserID)
	}
	for idx, note := range notes {
		lgr.WithField("idx", idx).WithField("note", note).Debug("notes")
	}

	c.pushNotifications(ctx, notes)

	session.MarkMessage(msg, "")
	lgr.WithField("msg", msg).Debug("marked")
	return nil
}

func (c *Consumer) pushNotifications(ctx context.Context, notifications []*alerts.Notification) {
	lgr := c.logger(ctx)

	// Notes could be pushed into a Kafka topic to have a more durable retry,
	// but that can be added later.
	for _, notification := range notifications {
		lgr := lgr.WithField("recipientUserID", notification.RecipientUserID)
		tokens, err := c.DeviceTokens.GetDeviceTokens(ctx, notification.RecipientUserID)
		if err != nil {
			lgr.WithError(err).Info("Unable to retrieve device tokens")
		}
		if len(tokens) == 0 {
			lgr.Debug("no device tokens found, won't push any notifications")
		}
		pushNote := alerts.ToPushNotification(notification)
		for _, token := range tokens {
			err := c.Pusher.Push(ctx, token, pushNote)
			if err != nil {
				lgr.WithError(err).Info("Unable to push notification")
			} else {
				notification.Sent(time.Now())
			}
		}
	}
}

// logger produces a log.Logger.
//
// It tries a number of options before falling back to a null Logger.
func (c *Consumer) logger(ctx context.Context) log.Logger {
	// A context's Logger is preferred, as it has the most... context.
	if ctxLgr := log.LoggerFromContext(ctx); ctxLgr != nil {
		return ctxLgr
	}
	if c.Logger != nil {
		return c.Logger
	}
	fallback, err := logjson.NewLogger(os.Stderr, log.DefaultLevelRanks(), log.DefaultLevel())
	if err != nil {
		fallback = lognull.NewLogger()
	}
	return fallback
}

type AlertsEvaluator interface {
	// EvaluateData to check if notifications should be sent in response to new data.
	EvaluateData(ctx context.Context, followedUserID, dataSetID string) ([]*alerts.Notification, error)
}

func unmarshalMessageValue[A any](b []byte, payload *A) ([]string, error) {
	wrapper := &struct {
		FullDocument      A `json:"fullDocument"`
		UpdateDescription struct {
			UpdatedFields map[string]any `json:"updatedFields"`
		} `json:"updateDescription"`
	}{}
	if err := bson.UnmarshalExtJSON(b, false, wrapper); err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal ExtJSON")
	}
	*payload = wrapper.FullDocument
	fields := []string{}
	for k := range wrapper.UpdateDescription.UpdatedFields {
		fields = append(fields, k)
	}
	return fields, nil
}

type AlertsClient interface {
	Delete(context.Context, *alerts.Config) error
	Get(context.Context, *alerts.Config) (*alerts.Config, error)
	List(_ context.Context, userID string) ([]*alerts.Config, error)
	Upsert(context.Context, *alerts.Config) error
}

// Pusher is a service-agnostic interface for sending push notifications.
type Pusher interface {
	// Push a notification to a device.
	Push(context.Context, *devicetokens.DeviceToken, *push.Notification) error
}
