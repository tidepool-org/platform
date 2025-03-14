package alerts

import (
	"context"
	"slices"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

const CarePartnerType = "org.tidepool.carepartner"

func NewCarePartnerTaskCreate() *task.TaskCreate {
	return &task.TaskCreate{
		Name:          pointer.FromAny(CarePartnerType),
		Type:          CarePartnerType,
		AvailableTime: &time.Time{},
		Data:          map[string]interface{}{},
	}
}

type CarePartnerRunner struct {
	logger log.Logger

	alerts       AlertsClient
	authClient   auth.ServerSessionTokenProvider
	deviceTokens auth.DeviceTokensClient
	permissions  permission.Client
	pusher       Pusher
}

// AlertsClient abstracts the alerts collection for the CarePartnerRunner.
//
// One implementation is [Client].
type AlertsClient interface {
	List(_ context.Context, followedUserID string) ([]*Config, error)
	Upsert(context.Context, *Config) error
	// OverdueCommunications returns a slice of [LastCommunication] for users that haven't
	// uploaded data recently.
	OverdueCommunications(context.Context) ([]LastCommunication, error)
}

func NewCarePartnerRunner(logger log.Logger, alerts AlertsClient,
	deviceTokens auth.DeviceTokensClient, pusher Pusher, permissions permission.Client,
	authClient auth.ServerSessionTokenProvider) (*CarePartnerRunner, error) {

	return &CarePartnerRunner{
		logger:       logger,
		alerts:       alerts,
		authClient:   authClient,
		deviceTokens: deviceTokens,
		pusher:       pusher,
		permissions:  permissions,
	}, nil
}

func (r *CarePartnerRunner) GetRunnerType() string {
	return CarePartnerType
}

func (r *CarePartnerRunner) GetRunnerTimeout() time.Duration {
	return r.GetRunnerDurationMaximum()
}

func (r *CarePartnerRunner) GetRunnerDeadline() time.Time {
	return time.Now().Add(3 * r.GetRunnerDurationMaximum())
}

const RunnerDurationMaximum = 30 * time.Second

func (r *CarePartnerRunner) GetRunnerDurationMaximum() time.Duration {
	return RunnerDurationMaximum
}

func (r *CarePartnerRunner) Run(ctx context.Context, tsk *task.Task) {
	r.logger.Info("care partner no communication check")
	start := time.Now()
	ctx = auth.NewContextWithServerSessionTokenProvider(ctx, r.authClient)
	if err := r.evaluateLastComms(ctx); err != nil {
		r.logger.WithError(err).Warn("running care partner no communication check")
	}
	tsk.RepeatAvailableAfter(time.Second - time.Since(start))
}

func (r *CarePartnerRunner) evaluateLastComms(ctx context.Context) error {
	overdue, err := r.alerts.OverdueCommunications(ctx)
	if err != nil {
		return errors.Wrap(err, "listing users without communication")
	}

	for _, lastComm := range overdue {
		if err := r.evaluateLastComm(ctx, lastComm); err != nil {
			r.logger.WithError(err).
				WithField("followedUserID", lastComm.UserID).
				WithField("dataSetID", lastComm.DataSetID).
				Info("Unable to evaluate no communication")
			continue
		}
	}

	return nil
}

func (r *CarePartnerRunner) evaluateLastComm(ctx context.Context,
	lastComm LastCommunication) error {

	configs, err := r.alerts.List(ctx, lastComm.UserID)
	if err != nil {
		return errors.Wrap(err, "listing follower alerts configs")
	}

	configs = slices.DeleteFunc(configs, r.authDenied(ctx))
	configs = slices.DeleteFunc(configs, func(config *Config) bool {
		return config.UploadID != lastComm.DataSetID
	})

	notifications := []*Notification{}
	for _, config := range configs {
		lgr := config.LoggerWithFields(r.logger)
		lastData := lastComm.LastReceivedDeviceData
		notification, needsUpsert := config.EvaluateNoCommunication(ctx, lgr, lastData)
		if notification != nil {
			notification.Sent = r.wrapWithUpsert(ctx, lgr, config, notification.Sent)
			notifications = append(notifications, notification)
		}
		if needsUpsert {
			err := r.alerts.Upsert(ctx, config)
			if err != nil {
				lgr.WithError(err).Error("Unable to upsert changed alerts config")
			}
		}
	}

	r.pushNotifications(ctx, notifications)

	return nil
}

// wrapWithUpsert to upsert the Config that triggered the Notification after it's sent.
func (r *CarePartnerRunner) wrapWithUpsert(ctx context.Context, lgr log.Logger, config *Config,
	original func(time.Time)) func(time.Time) {

	return func(at time.Time) {
		if original != nil {
			original(at)
		}
		if err := r.alerts.Upsert(ctx, config); err != nil {
			lgr.WithError(err).Error("Unable to upsert changed alerts config")
		}
	}
}

func (r *CarePartnerRunner) authDenied(ctx context.Context) func(*Config) bool {
	return func(c *Config) bool {
		if c == nil {
			return true
		}
		logger := r.logger.WithFields(log.Fields{
			"userID":         c.UserID,
			"followedUserID": c.FollowedUserID,
		})
		perms, err := r.permissions.GetUserPermissions(ctx, c.UserID, c.FollowedUserID)
		if err != nil {
			logger.WithError(err).Warn("Unable to confirm permissions; skipping")
			return true
		}
		if _, found := perms[permission.Follow]; !found {
			logger.Debug("permission denied: skipping")
			return true
		}
		return false
	}
}

func (r *CarePartnerRunner) pushNotifications(ctx context.Context,
	notifications []*Notification) {

	for _, notification := range notifications {
		lgr := r.logger.WithField("recipientUserID", notification.RecipientUserID)
		tokens, err := r.deviceTokens.GetDeviceTokens(ctx, notification.RecipientUserID)
		if err != nil {
			lgr.WithError(err).Info("unable to retrieve device tokens")
		}
		if len(tokens) == 0 {
			lgr.Debug("no device tokens found, won't push any notifications")
		}
		pushNotification := ToPushNotification(notification)
		for _, token := range tokens {
			err := r.pusher.Push(ctx, token, pushNotification)
			if err != nil {
				lgr.WithError(err).Info("unable to push notification")
			} else {
				notification.Sent(time.Now())
			}
		}
	}
}
