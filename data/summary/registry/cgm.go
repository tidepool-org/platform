package registry

import (
	"context"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/summary/store"
	"github.com/tidepool-org/platform/data/summary/types"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"time"
)

type CGMSummarizer struct {
	deviceData dataStore.DataRepository
	summaries  *store.Repo[types.CGMStats]
}

// Compile time interface check
var _ Summarizer[types.CGMStats] = &CGMSummarizer{}

func NewCGMSummarizer(collection *storeStructuredMongo.Repository, deviceData dataStore.DataRepository) Summarizer[types.CGMStats] {
	return &CGMSummarizer{
		deviceData: deviceData,
		summaries:  store.New[types.CGMStats](collection),
	}
}

func (c *CGMSummarizer) GetSummary(ctx context.Context, userId string) (*types.Summary[types.CGMStats], error) {
	return c.summaries.GetSummary(ctx, userId)
}

func (c *CGMSummarizer) UpdateSummary(ctx context.Context, userId string) (*types.Summary[types.CGMStats], error) {
	var status *types.UserLastUpdated
	var err error
	var userSummary *types.Summary[types.CGMStats]
	var userData []*glucoseDatum.Glucose

	timestamp := time.Now().UTC()
	logger := log.LoggerFromContext(ctx)
	userSummary, err = c.GetSummary(ctx, userId)
	if err != nil {
		return userSummary, err
	}

	logger.Debugf("Starting summary calculation for %s", userId)

	status, err = c.deviceData.GetLastUpdatedForUser(ctx, userId, userSummary.Type)
	if err != nil {
		return nil, err
	}

	// this filters out users which require no update, as they have no cgm data, but have an outdated summary
	if status.LastData.IsZero() {
		if userSummary != nil {
			// user's data is inactive/deleted, or this summary shouldn't have been created
			// TODO extract this into function and make all nil (maybe, previously moved away from pointers for easier code)
			userSummary.Dates.LastUpdatedDate = timestamp
			userSummary.Dates.OutdatedSince = nil
			userSummary.Dates.LastUploadDate = time.Time{}
			userSummary.Dates.LastData = nil
			userSummary.Dates.FirstData = time.Time{}
			logger.Warnf("User %s has an outdated summary with no data, skipping calc.", userId)
			userSummary, err = c.summaries.UpsertSummary(ctx, userSummary)
			if err != nil {
				return nil, err
			}
		}
		return userSummary, nil
	}

	// user exists (has relevant data), but no summary, create a blank one
	if userSummary == nil {
		userSummary = pointer.FromAny(types.Create[types.CGMStats](userId))
	}

	// remove 30 days for start time
	startTime := status.LastData.AddDate(0, 0, -30)
	endTime := status.LastData

	if userSummary.Dates.LastData != nil {
		// if summary already exists with a last data checkpoint, start data pull there
		if startTime.Before(*userSummary.Dates.LastData) {
			startTime = *userSummary.Dates.LastData
		}

		// ensure endTime does not move backwards by capping it at summary LastData
		if !status.LastData.After(*userSummary.Dates.LastData) {
			endTime = *userSummary.Dates.LastData
		}
	}

	err = c.deviceData.GetDataRange(ctx, userData, userId, userSummary.Type, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// skip past data
	if len(userSummary.Stats.HourlyStats) > 0 {
		userData, err = SkipUntil(userSummary.Stats.HourlyStats[len(userSummary.Stats.HourlyStats)-1].Date, userData)
	}

	// if there is new data
	if len(userData) > 0 {
		err = types.Update(userSummary.Stats, userData)
		if err != nil {
			return nil, err
		}
	} else {
		// "new" data must be in the past, don't update, just remove flags and set new date
		logger.Infof("User %s has an outdated summary with no forward data, skipping calc.", userId)
	}

	userSummary.Dates.LastUpdatedDate = timestamp
	userSummary.Dates.OutdatedSince = nil
	userSummary.Dates.LastUploadDate = status.LastUpload
	userSummary.Dates.LastData = userData[len(userData)].Time
	userSummary.Dates.FirstData = userSummary.Stats.HourlyStats[0].Date

	// technically, this never could be zero, but we check anyway
	userSummary.Dates.HasLastUploadDate = !status.LastUpload.IsZero()

	userSummary, err = c.summaries.UpsertSummary(ctx, userSummary)

	return userSummary, err
}
