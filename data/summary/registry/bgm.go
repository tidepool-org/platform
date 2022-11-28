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

const (
	backfillBatch = 100000
)

type BGMSummarizer struct {
	deviceData dataStore.DataRepository
	summaries  *store.Repo[types.BGMStats]
}

// Compile time interface check
var _ Summarizer[types.BGMStats] = &BGMSummarizer{}

func NewBGMSummarizer(collection *storeStructuredMongo.Repository, deviceData dataStore.DataRepository) Summarizer[types.BGMStats] {
	return &BGMSummarizer{
		deviceData: deviceData,
		summaries:  store.New[types.BGMStats](collection),
	}
}

func (c *BGMSummarizer) GetSummary(ctx context.Context, userId string) (*types.Summary[types.BGMStats], error) {
	return c.summaries.GetSummary(ctx, userId)
}

func (c *BGMSummarizer) UpdateSummary(ctx context.Context, userId string) (*types.Summary[types.BGMStats], error) {
	var status *types.UserLastUpdated
	var err error
	var userSummary *types.Summary[types.BGMStats]
	var userData []*glucoseDatum.Glucose

	timestamp := time.Now().UTC()
	logger := log.LoggerFromContext(ctx)
	userSummary, err = c.GetSummary(ctx, userId)
	if err != nil {
		return userSummary, err
	}

	logger.Debugf("Starting summary calculation for %s", userId)

	status, err = c.deviceData.GetLastUpdatedForUser(ctx, userId, "bgm")
	if err != nil {
		return userSummary, err
	}

	// this filters out users which require no update, as they have no cgm data, but have an outdated summary
	if status.LastData.IsZero() {
		if userSummary != nil {
			// user's data is inactive/deleted, or this summary shouldn't have been created
			logger.Warnf("User %s has an outdated summary with no data, skipping calc.", userId)
			userSummary, err = c.summaries.UpsertSummary(ctx, userSummary)
			if err != nil {
				return userSummary, err
			}
		}
		return userSummary, nil
	}

	// user exists (has relevant data), but no summary, create a blank one
	if userSummary == nil {
		userSummary = pointer.FromAny(types.Create[types.BGMStats](userId))
	}

	if !status.LastData.IsZero() {
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

		userData, err = c.deviceData.GetDataRange(ctx, userId, "cgm", startTime, endTime)
		if err != nil {
			return nil, err
		}
	}

	// if there is new data
	if len(userData) > 0 {
		err = userSummary.Stats.CalculateStats(userData)
		if err != nil {
			return userSummary, err
		}

		userSummary.Stats.CalculateSummary()
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

	//status *UserLastUpdated, userBGMData []*glucoseDatum.Glucose)

	return userSummary, nil
}

func (c *BGMSummarizer) BackfillSummaries(ctx context.Context) (int, error) {
	var empty struct{}
	var userIDsReqBackfill []string
	var count = 0

	distinctDataUserIDs, err := c.deviceData.DistinctUserIDs(ctx)
	if err != nil {
		return count, err
	}

	distinctSummaryIDs, err := c.summaries.DistinctSummaryIDs(ctx)
	if err != nil {
		return count, err
	}

	distinctSummaryIDMap := make(map[string]struct{})
	for _, v := range distinctSummaryIDs {
		distinctSummaryIDMap[v] = empty
	}

	for _, userID := range distinctDataUserIDs {
		if _, exists := distinctSummaryIDMap[userID]; exists {
		} else {
			userIDsReqBackfill = append(userIDsReqBackfill, userID)
		}

		if len(userIDsReqBackfill) >= backfillBatch {
			break
		}
	}

	var summaries = make([]*types.Summary[types.BGMStats], len(userIDsReqBackfill))
	for i, userId := range userIDsReqBackfill {
		summaries[i] = pointer.FromAny(types.Create[types.BGMStats](userId))
	}

	if len(summaries) > 0 {
		count, err = c.summaries.CreateSummaries(ctx, summaries)
		if err != nil {
			return count, err
		}
	}

	return count, nil
}
