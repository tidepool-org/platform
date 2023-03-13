package summary

import (
	"context"
	"time"

	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/summary/store"
	"github.com/tidepool-org/platform/data/summary/types"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

const (
	backfillBatch = 100000
)

type SummarizerRegistry struct {
	summarizers map[string]any
}

func New(summaryRepository *storeStructuredMongo.Repository, dataRepository dataStore.DataRepository) *SummarizerRegistry {
	registry := &SummarizerRegistry{summarizers: make(map[string]any)}
	addSummarizer(registry, NewBGMSummarizer(summaryRepository, dataRepository))
	addSummarizer(registry, NewCGMSummarizer(summaryRepository, dataRepository))
	return registry
}

func addSummarizer[T types.Stats, A types.StatsPt[T]](reg *SummarizerRegistry, summarizer Summarizer[T, A]) {
	typ := types.GetTypeString[T, A]()
	reg.summarizers[typ] = summarizer
}

func GetSummarizer[T types.Stats, A types.StatsPt[T]](reg *SummarizerRegistry) Summarizer[T, A] {
	typ := types.GetTypeString[T, A]()
	summarizer := reg.summarizers[typ]
	return summarizer.(Summarizer[T, A])
}

type Summarizer[T types.Stats, A types.StatsPt[T]] interface {
	GetSummary(ctx context.Context, userId string) (*types.Summary[T, A], error)
	SetOutdated(ctx context.Context, userId string) error
	UpdateSummary(ctx context.Context, userId string) (*types.Summary[T, A], error)
	GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error)
	BackfillSummaries(ctx context.Context) (int, error)
}

// Compile time interface check
var _ Summarizer[types.CGMStats, *types.CGMStats] = &GlucoseSummarizer[types.CGMStats, *types.CGMStats]{}
var _ Summarizer[types.BGMStats, *types.BGMStats] = &GlucoseSummarizer[types.BGMStats, *types.BGMStats]{}

type GlucoseSummarizer[T types.Stats, A types.StatsPt[T]] struct {
	deviceData dataStore.DataRepository
	summaries  *store.Repo[T, A]
}

func NewBGMSummarizer(collection *storeStructuredMongo.Repository, deviceData dataStore.DataRepository) Summarizer[types.BGMStats, *types.BGMStats] {
	return &GlucoseSummarizer[types.BGMStats, *types.BGMStats]{
		deviceData: deviceData,
		summaries:  store.New[types.BGMStats, *types.BGMStats](collection),
	}
}

func NewCGMSummarizer(collection *storeStructuredMongo.Repository, deviceData dataStore.DataRepository) Summarizer[types.CGMStats, *types.CGMStats] {
	return &GlucoseSummarizer[types.CGMStats, *types.CGMStats]{
		deviceData: deviceData,
		summaries:  store.New[types.CGMStats, *types.CGMStats](collection),
	}
}

func (c *GlucoseSummarizer[T, A]) GetSummary(ctx context.Context, userId string) (*types.Summary[T, A], error) {
	return c.summaries.GetSummary(ctx, userId)
}

func (c *GlucoseSummarizer[T, A]) SetOutdated(ctx context.Context, userId string) error {
	_, err := c.summaries.SetOutdated(ctx, userId)
	return err
}

func (c *GlucoseSummarizer[T, A]) GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error) {
	return c.summaries.GetOutdatedUserIDs(ctx, pagination)
}

func (c *GlucoseSummarizer[T, A]) BackfillSummaries(ctx context.Context) (int, error) {
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

	var summaries = make([]*types.Summary[T, A], len(userIDsReqBackfill))
	for i, userID := range userIDsReqBackfill {
		summaries[i] = types.Create[T, A](userID)
	}

	if len(summaries) > 0 {
		count, err = c.summaries.CreateSummaries(ctx, summaries)
		if err != nil {
			return count, err
		}
	}

	return count, nil
}

func (c *GlucoseSummarizer[T, A]) UpdateSummary(ctx context.Context, userId string) (*types.Summary[T, A], error) {
	var status *types.UserLastUpdated
	var err error
	var userSummary *types.Summary[T, A]
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
		userSummary = types.Create[T, A](userId)
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
	bucketsLen := userSummary.Stats.GetBucketsLen()
	if bucketsLen > 0 {
		userData, err = types.SkipUntil(userSummary.Stats.GetBucketDate(bucketsLen-1), userData)
	}

	// if there is new data
	if len(userData) > 0 {
		err = userSummary.Stats.Update(userData)
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
	userSummary.Dates.FirstData = userSummary.Stats.GetBucketDate(0)

	// technically, this never could be zero, but we check anyway
	userSummary.Dates.HasLastUploadDate = !status.LastUpload.IsZero()

	userSummary, err = c.summaries.UpsertSummary(ctx, userSummary)

	return userSummary, err
}
