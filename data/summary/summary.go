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
	backfillBatch       = 50000
	backfillInsertBatch = 10000
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
	SetOutdated(ctx context.Context, userId string) (*time.Time, error)
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

func (c *GlucoseSummarizer[T, A]) SetOutdated(ctx context.Context, userId string) (*time.Time, error) {
	return c.summaries.SetOutdated(ctx, userId)
}

func (c *GlucoseSummarizer[T, A]) GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error) {
	return c.summaries.GetOutdatedUserIDs(ctx, pagination)
}

func (c *GlucoseSummarizer[T, A]) BackfillSummaries(ctx context.Context) (int, error) {
	var empty struct{}

	distinctDataUserIDs, err := c.deviceData.DistinctUserIDs(ctx, types.GetDeviceDataTypeString[T, A]())
	if err != nil {
		return 0, err
	}

	distinctSummaryIDs, err := c.summaries.DistinctSummaryIDs(ctx)
	if err != nil {
		return 0, err
	}

	distinctSummaryIDMap := make(map[string]struct{})
	for _, v := range distinctSummaryIDs {
		distinctSummaryIDMap[v] = empty
	}

	userIDsReqBackfill := make([]string, 0, backfillBatch)
	for _, userID := range distinctDataUserIDs {
		if _, exists := distinctSummaryIDMap[userID]; !exists {
			userIDsReqBackfill = append(userIDsReqBackfill, userID)
		}

		if len(userIDsReqBackfill) >= backfillBatch {
			break
		}
	}

	summaries := make([]*types.Summary[T, A], 0, len(userIDsReqBackfill))
	for _, userID := range userIDsReqBackfill {
		s := types.Create[T, A](userID)
		s.SetOutdated()
		summaries = append(summaries, s)

		if len(summaries) >= backfillInsertBatch {
			count, err := c.summaries.CreateSummaries(ctx, summaries)
			if err != nil {
				return count, err
			}
			summaries = nil
		}
	}

	if len(summaries) > 0 {
		return c.summaries.CreateSummaries(ctx, summaries)
	}

	return 0, nil
}

func (c *GlucoseSummarizer[T, A]) UpdateSummary(ctx context.Context, userId string) (*types.Summary[T, A], error) {
	logger := log.LoggerFromContext(ctx)
	userSummary, err := c.GetSummary(ctx, userId)
	if err != nil {
		return userSummary, err
	}

	logger.Debugf("Starting summary calculation for %s", userId)

	status, err := c.deviceData.GetLastUpdatedForUser(ctx, userId, types.GetDeviceDataTypeString[T, A]())
	if err != nil {
		return nil, err
	}

	// this filters out users which require no update, as they have no data of type T, but have an outdated summary
	if status.LastData.IsZero() {
		if userSummary != nil {
			// user's data is inactive/deleted, or this summary shouldn't have been created
			userSummary.Dates.ZeroOut()
			logger.Warnf("User %s has an outdated summary with no data, skipping calc.", userId)
			err = c.summaries.UpsertSummary(ctx, userSummary)
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

	if userSummary.Dates.LastData != nil {
		// if summary already exists with a last data checkpoint, start data pull there
		if startTime.Before(*userSummary.Dates.LastData) {
			startTime = *userSummary.Dates.LastData
		}

		// ensure LastData does not move backwards by capping it at summary LastData
		if status.LastData.Before(*userSummary.Dates.LastData) {
			status.LastData = *userSummary.Dates.LastData
		}
	}

	var userData []*glucoseDatum.Glucose
	err = c.deviceData.GetDataRange(ctx, &userData, userId, types.GetDeviceDataTypeString[T, A](), startTime, status.LastData)
	if err != nil {
		return nil, err
	}

	// skip past data
	bucketsLen := userSummary.Stats.GetBucketsLen()
	if bucketsLen > 0 {
		userData, err = types.SkipUntil(userSummary.Stats.GetBucketDate(bucketsLen-1), userData)
	}

	// if there is no new data
	if len(userData) < 0 {
		userSummary.UpdateWithoutChangeCount++
		logger.Infof("User %s has an outdated summary with no forward data, summary will not be calculated.", userId)
	}

	err = userSummary.Stats.Update(userData)
	if err != nil {
		return nil, err
	}

	userSummary.Dates.Update(status, userSummary.Stats.GetBucketDate(0))

	err = c.summaries.UpsertSummary(ctx, userSummary)

	return userSummary, err
}
