package events

import (
	"cmp"
	"context"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/data/store"
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
	Alerts       AlertsClient
	Data         store.DataRepository
	DeviceTokens auth.DeviceTokensClient
	Evaluator    AlertsEvaluator
	Permissions  permission.Client
	Pusher       Pusher
	Tokens       alerts.TokenProvider

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
	if err := unmarshalMessageValue(msg.Value, cfg); err != nil {
		return err
	}
	lgr := c.logger(ctx)
	lgr.WithField("cfg", cfg).Info("consuming an alerts config message")

	ctxLog := c.logger(ctx).WithField("followedUserID", cfg.FollowedUserID)
	ctx = log.NewContextWithLogger(ctx, ctxLog)

	notes, err := c.Evaluator.Evaluate(ctx, cfg.FollowedUserID)
	if err != nil {
		format := "Unable to evalaute alerts configs triggered event for user %s"
		return errors.Wrapf(err, format, cfg.UserID)
	}
	ctxLog.WithField("notes", notes).Debug("notes generated from alerts config")

	c.pushNotes(ctx, notes)

	session.MarkMessage(msg, "")
	lgr.WithField("message", msg).Debug("marked")
	return nil
}

func (c *Consumer) consumeDeviceData(ctx context.Context,
	session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) (err error) {

	datum := &Glucose{}
	if err := unmarshalMessageValue(msg.Value, datum); err != nil {
		return err
	}
	lgr := c.logger(ctx)
	lgr.WithField("data", datum).Info("consuming a device data message")

	if datum.UserID == nil {
		return errors.New("Unable to retrieve alerts configs: userID is nil")
	}
	ctx = log.NewContextWithLogger(ctx, lgr.WithField("followedUserID", *datum.UserID))
	notes, err := c.Evaluator.Evaluate(ctx, *datum.UserID)
	if err != nil {
		format := "Unable to evalaute device data triggered event for user %s"
		return errors.Wrapf(err, format, *datum.UserID)
	}
	for idx, note := range notes {
		lgr.WithField("idx", idx).WithField("note", note).Debug("notes")
	}

	c.pushNotes(ctx, notes)

	session.MarkMessage(msg, "")
	lgr.WithField("message", msg).Debug("marked")
	return nil
}

func (c *Consumer) pushNotes(ctx context.Context, notifications []*alerts.Notification) {
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
		pushNote := push.FromAlertsNotification(notification)
		for _, token := range tokens {
			if err := c.Pusher.Push(ctx, token, pushNote); err != nil {
				lgr.WithError(err).Info("Unable to push notification")
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
	Evaluate(ctx context.Context, followedUserID string) ([]*alerts.Notification, error)
}

func NewAlertsEvaluator(alerts AlertsClient, data store.DataRepository,
	perms permission.Client, tokens alerts.TokenProvider) *evaluator {

	return &evaluator{
		Alerts:      alerts,
		Data:        data,
		Permissions: perms,
		Tokens:      tokens,
	}
}

// evaluator implements AlertsEvaluator.
type evaluator struct {
	Alerts      AlertsClient
	Data        store.DataRepository
	Permissions permission.Client
	Tokens      alerts.TokenProvider
}

// logger produces a log.Logger.
//
// It tries a number of options before falling back to a null Logger.
func (e *evaluator) logger(ctx context.Context) log.Logger {
	// A context's Logger is preferred, as it has the most... context.
	if ctxLgr := log.LoggerFromContext(ctx); ctxLgr != nil {
		return ctxLgr
	}
	fallback, err := logjson.NewLogger(os.Stderr, log.DefaultLevelRanks(), log.DefaultLevel())
	if err != nil {
		fallback = lognull.NewLogger()
	}
	return fallback
}

// Evaluate followers' alerts.Configs to generate alert notifications.
func (e *evaluator) Evaluate(ctx context.Context, followedUserID string) (
	[]*alerts.Notification, error) {

	alertsConfigs, err := e.gatherAlertsConfigs(ctx, followedUserID)
	if err != nil {
		return nil, err
	}

	alertsConfigsByUploadID := e.mapAlertsConfigsByUploadID(alertsConfigs)

	notifications := []*alerts.Notification{}
	for uploadID, cfgs := range alertsConfigsByUploadID {
		resp, err := e.gatherData(ctx, followedUserID, uploadID, cfgs)
		if err != nil {
			return nil, err
		}
		notifications = slices.Concat(notifications, e.generateNotes(ctx, cfgs, resp))
	}

	return notifications, nil
}

func (e *evaluator) mapAlertsConfigsByUploadID(cfgs []*alerts.Config) map[string][]*alerts.Config {
	mapped := map[string][]*alerts.Config{}
	for _, cfg := range cfgs {
		if _, found := mapped[cfg.UploadID]; !found {
			mapped[cfg.UploadID] = []*alerts.Config{}
		}
		mapped[cfg.UploadID] = append(mapped[cfg.UploadID], cfg)
	}
	return mapped
}

func (e *evaluator) gatherAlertsConfigs(ctx context.Context,
	followedUserID string) ([]*alerts.Config, error) {

	alertsConfigs, err := e.Alerts.List(ctx, followedUserID)
	if err != nil {
		return nil, err
	}
	alertsConfigs = slices.DeleteFunc(alertsConfigs, e.authDenied(ctx))
	return alertsConfigs, nil
}

// authDenied builds functions that enable slices.DeleteFunc to remove
// unauthorized users' alerts.Configs.
//
// Via a closure it's able to inject information from the Context and the
// evaluator itself into the resulting function.
func (e *evaluator) authDenied(ctx context.Context) func(ac *alerts.Config) bool {
	lgr := e.logger(ctx)
	return func(ac *alerts.Config) bool {
		if ac == nil {
			return true
		}
		lgr = lgr.WithFields(log.Fields{
			"userID":         ac.UserID,
			"followedUserID": ac.FollowedUserID,
		})
		token, err := e.Tokens.ServerSessionToken()
		if err != nil {
			lgr.WithError(err).Warn("Unable to confirm permissions; skipping")
			return false
		}
		ctx = auth.NewContextWithServerSessionToken(ctx, token)
		perms, err := e.Permissions.GetUserPermissions(ctx, ac.UserID, ac.FollowedUserID)
		if err != nil {
			lgr.WithError(err).Warn("Unable to confirm permissions; skipping")
			return true
		}
		if _, found := perms[permission.Follow]; !found {
			lgr.Debug("permission denied: skipping")
			return true
		}
		return false
	}
}

func (e *evaluator) gatherData(ctx context.Context, followedUserID, uploadID string,
	alertsConfigs []*alerts.Config) (*store.AlertableResponse, error) {

	if len(alertsConfigs) == 0 {
		return nil, nil
	}

	longestDelay := slices.MaxFunc(alertsConfigs, func(i, j *alerts.Config) int {
		return cmp.Compare(i.LongestDelay(), j.LongestDelay())
	}).LongestDelay()
	longestDelay = max(5*time.Minute, longestDelay)
	params := store.AlertableParams{
		UserID:   followedUserID,
		UploadID: uploadID,
		Start:    time.Now().Add(-longestDelay),
	}
	resp, err := e.Data.GetAlertableData(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (e *evaluator) generateNotes(ctx context.Context,
	alertsConfigs []*alerts.Config, resp *store.AlertableResponse) []*alerts.Notification {

	if len(alertsConfigs) == 0 {
		return nil
	}

	lgr := e.logger(ctx)
	notifications := []*alerts.Notification{}
	for _, alertsConfig := range alertsConfigs {
		l := lgr.WithFields(log.Fields{
			"userID":         alertsConfig.UserID,
			"followedUserID": alertsConfig.FollowedUserID,
			"uploadID":       alertsConfig.UploadID,
		})
		c := log.NewContextWithLogger(ctx, l)
		note := alertsConfig.Evaluate(c, resp.Glucose, resp.DosingDecisions)
		if note != nil {
			notifications = append(notifications, note)
			continue
		}
	}

	return notifications
}

func unmarshalMessageValue[A any](b []byte, payload *A) error {
	wrapper := &struct {
		FullDocument A `json:"fullDocument"`
	}{}
	if err := bson.UnmarshalExtJSON(b, false, wrapper); err != nil {
		return errors.Wrap(err, "Unable to unmarshal ExtJSON")
	}
	*payload = wrapper.FullDocument
	return nil
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
