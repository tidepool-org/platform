package alerts

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/tidepool-org/platform/auth"
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
	Alerts        Repository
	Data          DataRepository
	Logger        log.Logger
	Permissions   permission.Client
	TokenProvider auth.ServerSessionTokenProvider
}

func NewEvaluator(alerts Repository, dataRepo DataRepository, permissions permission.Client,
	logger log.Logger, tokenProvider auth.ServerSessionTokenProvider) *Evaluator {

	return &Evaluator{
		Alerts:        alerts,
		Data:          dataRepo,
		Logger:        logger,
		Permissions:   permissions,
		TokenProvider: tokenProvider,
	}
}

// EvaluateData generates alert notifications in response to a user uploading data.
func (e *Evaluator) EvaluateData(ctx context.Context, followedUserID, dataSetID string) (
	[]*Notification, error) {

	configs, err := e.gatherConfigs(ctx, followedUserID, dataSetID)
	if err != nil {
		return nil, err
	}

	configsByDataSetID := e.mapConfigsByDataSetID(configs)

	notifications := []*Notification{}
	for dsID, configs := range configsByDataSetID {
		resp, err := e.gatherData(ctx, followedUserID, dsID, configs)
		if err != nil {
			return nil, err
		}
		for _, config := range configs {
			lgr := config.LoggerWithFields(e.Logger)
			notification, needsUpsert := e.genNotificationForConfig(ctx, lgr, config, resp)
			if notification != nil {
				notifications = append(notifications, notification)
			}
			if needsUpsert {
				err := e.Alerts.Upsert(ctx, config)
				if err != nil {
					lgr.WithError(err).Error("Unable to upsert changed alerts config")
				}
			}
		}
	}

	return notifications, nil
}

func (e *Evaluator) genNotificationForConfig(ctx context.Context, lgr log.Logger,
	config *Config, resp *GetAlertableDataResponse) (*Notification, bool) {

	notification, needsUpsert := config.EvaluateData(ctx, resp.Glucose, resp.DosingDecisions)
	if notification != nil {
		notification.Sent = e.wrapWithUpsert(ctx, lgr, config, notification.Sent)
	}
	return notification, needsUpsert
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
		ctx = auth.NewContextWithServerSessionTokenProvider(ctx, e.TokenProvider)
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

	resp.Glucose = slices.DeleteFunc(resp.Glucose,
		func(g *glucose.Glucose) bool { return g.Time == nil })
	resp.DosingDecisions = slices.DeleteFunc(resp.DosingDecisions,
		func(d *dosingdecision.DosingDecision) bool { return d.Time == nil })

	return resp, nil
}

// wrapWithUpsert to upsert the Config that triggered the Notification after it's sent.
func (e *Evaluator) wrapWithUpsert(ctx context.Context, lgr log.Logger, config *Config,
	original func(time.Time)) func(time.Time) {

	return func(at time.Time) {
		if original != nil {
			original(at)
		}
		if err := e.Alerts.Upsert(ctx, config); err != nil {
			lgr.WithError(err).Error("Unable to upsert changed alerts config")
		}
	}
}
