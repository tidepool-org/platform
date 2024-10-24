package alerts

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
)

// DataRepository encapsulates queries of the data collection for use with alerts.
type DataRepository interface {
	// GetAlertableData queries for the data used to evaluate alerts configurations.
	GetAlertableData(ctx context.Context, params GetAlertableDataParams) (*GetAlertableDataResponse, error)
}

type GetAlertableDataParams struct {
	// UserID of the user that owns the data.
	UserID string
	// UploadID of the device data set to query.
	//
	// The term DataSetID should be preferred, but UploadID already existed in some places.
	UploadID string
	// Start limits the data to those recorded after this time.
	Start time.Time
	// End limits the data to those recorded before this time.
	End time.Time
}

type GetAlertableDataResponse struct {
	DosingDecisions []*dosingdecision.DosingDecision
	Glucose         []*glucose.Glucose
}

type Evaluator struct {
	Alerts      Repository
	Data        DataRepository
	Logger      log.Logger
	Permissions permission.Client
}

func NewEvaluator(alerts Repository, dataRepo DataRepository, permissions permission.Client,
	logger log.Logger) *Evaluator {

	return &Evaluator{
		Alerts:      alerts,
		Data:        dataRepo,
		Logger:      logger,
		Permissions: permissions,
	}
}

// EvaluateData generates alert notifications in response to a user uploading data.
func (e *Evaluator) EvaluateData(ctx context.Context, followedUserID, dataSetID string) (
	[]*NotificationWithHook, error) {

	configs, err := e.gatherConfigs(ctx, followedUserID, dataSetID)
	if err != nil {
		return nil, err
	}

	configsByDataSetID := e.mapConfigsByDataSetID(configs)

	notifications := []*NotificationWithHook{}
	for dsID, cfgs := range configsByDataSetID {
		resp, err := e.gatherData(ctx, followedUserID, dsID, cfgs)
		if err != nil {
			return nil, err
		}
		notifications = slices.Concat(notifications, e.generateNotes(ctx, cfgs, resp))
	}

	return notifications, nil
}

func (e *Evaluator) mapConfigsByDataSetID(cfgs []*Config) map[string][]*Config {
	mapped := map[string][]*Config{}
	for _, cfg := range cfgs {
		if _, found := mapped[cfg.UploadID]; !found {
			mapped[cfg.UploadID] = []*Config{}
		}
		mapped[cfg.UploadID] = append(mapped[cfg.UploadID], cfg)
	}
	return mapped
}

func (e *Evaluator) gatherConfigs(ctx context.Context, followedUserID, dataSetID string) (
	[]*Config, error) {

	configs, err := e.Alerts.List(ctx, followedUserID)
	if err != nil {
		return nil, err
	}
	configs = slices.DeleteFunc(configs, e.authDenied(ctx))
	configs = slices.DeleteFunc(configs, func(config *Config) bool {
		return config.UploadID != dataSetID
	})
	return configs, nil
}

// authDenied builds a function for slices.DeleteFunc to remove unauthorized users' Configs.
//
// This would catch the unintended case where a follower's permission was revoked, but their
// [Config] wasn't deleted.
//
// A closure is used to inject information from the evaluator into the resulting function.
func (e *Evaluator) authDenied(ctx context.Context) func(*Config) bool {
	return func(c *Config) bool {
		if c == nil {
			return true
		}
		logger := e.Logger.WithFields(log.Fields{
			"userID":         c.UserID,
			"followedUserID": c.FollowedUserID,
		})
		perms, err := e.Permissions.GetUserPermissions(ctx, c.UserID, c.FollowedUserID)
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

func (e *Evaluator) gatherData(ctx context.Context, followedUserID, dataSetID string,
	configs []*Config) (*GetAlertableDataResponse, error) {

	if len(configs) == 0 {
		return nil, nil
	}

	longestDelay := slices.MaxFunc(configs, func(i, j *Config) int {
		return cmp.Compare(i.LongestDelay(), j.LongestDelay())
	}).LongestDelay()
	longestDelay = max(5*time.Minute, longestDelay)
	params := GetAlertableDataParams{
		UserID:   followedUserID,
		UploadID: dataSetID,
		Start:    time.Now().Add(-longestDelay),
	}
	resp, err := e.Data.GetAlertableData(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (e *Evaluator) generateNotes(ctx context.Context, configs []*Config,
	resp *GetAlertableDataResponse) []*NotificationWithHook {

	if len(configs) == 0 {
		return nil
	}

	notifications := []*NotificationWithHook{}
	for _, config := range configs {
		lgr := e.Logger.WithFields(log.Fields{
			"userID":         config.UserID,
			"followedUserID": config.FollowedUserID,
			"uploadID":       config.UploadID,
		})
		evalCtx := log.NewContextWithLogger(ctx, lgr)
		notification, changed := config.EvaluateData(evalCtx, resp.Glucose, resp.DosingDecisions)
		if notification != nil {
			if notification.Sent != nil {
				notification.Sent = e.wrapWithUpsert(evalCtx, lgr, config, notification.Sent)
			}
			notifications = append(notifications, notification)
			continue
		} else if changed {
			// No notification was generated, so no further changes are expected. However,
			// there were activity changes that need persisting.
			err := e.Alerts.Upsert(ctx, config)
			if err != nil {
				lgr.WithError(err).Error("Unable to save changed alerts config")
				continue
			}
		}
	}

	return notifications
}

// wrapWithUpsert to upsert the Config that triggered the Notification after it's sent.
func (e *Evaluator) wrapWithUpsert(ctx context.Context,
	lgr log.Logger, config *Config, original SentFunc) SentFunc {

	return func(at time.Time) {
		original(at)
		if err := e.Alerts.Upsert(ctx, config); err != nil {
			lgr.WithError(err).Error("Unable to save changed alerts config")
		}
	}
}
