package summary

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/summary/fetcher"
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

func New(summaryRepository *storeStructuredMongo.Repository, bucketsRepository *storeStructuredMongo.Repository, fetcher fetcher.DeviceDataFetcher) *SummarizerRegistry {
	registry := &SummarizerRegistry{summarizers: make(map[string]any)}
	addSummarizer(registry, NewCGMSummarizer(summaryRepository, bucketsRepository, fetcher))
	addSummarizer(registry, NewBGMSummarizer(summaryRepository, bucketsRepository, fetcher))
	addSummarizer(registry, NewContinuousSummarizer(summaryRepository, bucketsRepository, fetcher))
	return registry
}

func addSummarizer[A types.StatsPt[T, P, B], P types.BucketDataPt[B], T types.Stats, B types.BucketData](reg *SummarizerRegistry, summarizer Summarizer[A, P, T, B]) {
	typ := types.GetTypeString[A, P]()
	reg.summarizers[typ] = summarizer
}

func GetSummarizer[A types.StatsPt[T, P, B], P types.BucketDataPt[B], T types.Stats, B types.BucketData](reg *SummarizerRegistry) Summarizer[A, P, T, B] {
	typ := types.GetTypeString[A, P]()
	summarizer := reg.summarizers[typ]
	return summarizer.(Summarizer[A, P, T, B])
}

type Summarizer[A types.StatsPt[T, P, B], P types.BucketDataPt[B], T types.Stats, B types.BucketData] interface {
	GetSummary(ctx context.Context, userId string) (*types.Summary[A, P, T, B], error)
	GetBucketsRange(ctx context.Context, userId string, startTime time.Time, endTime time.Time) (fetcher.AnyCursor, error)
	SetOutdated(ctx context.Context, userId, reason string) (*time.Time, error)
	UpdateSummary(ctx context.Context, userId string) (*types.Summary[A, P, T, B], error)
	GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error)
	GetMigratableUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error)
	BackfillSummaries(ctx context.Context) (int, error)
}

// Compile time interface check
var _ Summarizer[*types.CGMStats, *types.GlucoseBucket, types.CGMStats, types.GlucoseBucket] = &GlucoseSummarizer[*types.CGMStats, *types.GlucoseBucket, types.CGMStats, types.GlucoseBucket]{}
var _ Summarizer[*types.BGMStats, *types.GlucoseBucket, types.BGMStats, types.GlucoseBucket] = &GlucoseSummarizer[*types.BGMStats, *types.GlucoseBucket, types.BGMStats, types.GlucoseBucket]{}
var _ Summarizer[*types.ContinuousStats, *types.ContinuousBucket, types.ContinuousStats, types.ContinuousBucket] = &GlucoseSummarizer[*types.ContinuousStats, *types.ContinuousBucket, types.ContinuousStats, types.ContinuousBucket]{}

func CreateGlucoseDatum() data.Datum {
	return &glucoseDatum.Glucose{}
}

type GlucoseSummarizer[A types.StatsPt[T, P, B], P types.BucketDataPt[B], T types.Stats, B types.BucketData] struct {
	cursorFactory fetcher.DataCursorFactory
	dataFetcher   fetcher.DeviceDataFetcher
	summaries     *store.Summaries[A, P, T, B]
	buckets       *store.Buckets[P, B]
}

func NewBGMSummarizer(collection *storeStructuredMongo.Repository, bucketsCollection *storeStructuredMongo.Repository, dataFetcher fetcher.DeviceDataFetcher) Summarizer[*types.BGMStats, *types.GlucoseBucket, types.BGMStats, types.GlucoseBucket] {
	return &GlucoseSummarizer[*types.BGMStats, *types.GlucoseBucket, types.BGMStats, types.GlucoseBucket]{
		cursorFactory: func(c *mongo.Cursor) fetcher.DeviceDataCursor {
			return fetcher.NewDefaultCursor(c, CreateGlucoseDatum)
		},
		dataFetcher: dataFetcher,
		summaries:   store.NewSummaries[*types.BGMStats, *types.GlucoseBucket](collection),
		buckets:     store.NewBuckets[*types.GlucoseBucket](bucketsCollection, types.SummaryTypeBGM),
	}
}

func NewCGMSummarizer(collection *storeStructuredMongo.Repository, bucketsCollection *storeStructuredMongo.Repository, dataFetcher fetcher.DeviceDataFetcher) Summarizer[*types.CGMStats, *types.GlucoseBucket, types.CGMStats, types.GlucoseBucket] {
	return &GlucoseSummarizer[*types.CGMStats, *types.GlucoseBucket, types.CGMStats, types.GlucoseBucket]{
		cursorFactory: func(c *mongo.Cursor) fetcher.DeviceDataCursor {
			return fetcher.NewDefaultCursor(c, CreateGlucoseDatum)
		},
		dataFetcher: dataFetcher,
		summaries:   store.NewSummaries[*types.CGMStats, *types.GlucoseBucket](collection),
		buckets:     store.NewBuckets[*types.GlucoseBucket](bucketsCollection, types.SummaryTypeCGM),
	}
}

func NewContinuousSummarizer(collection *storeStructuredMongo.Repository, bucketsCollection *storeStructuredMongo.Repository, dataFetcher fetcher.DeviceDataFetcher) Summarizer[*types.ContinuousStats, *types.ContinuousBucket, types.ContinuousStats, types.ContinuousBucket] {
	return &GlucoseSummarizer[*types.ContinuousStats, *types.ContinuousBucket, types.ContinuousStats, types.ContinuousBucket]{
		cursorFactory: func(c *mongo.Cursor) fetcher.DeviceDataCursor {
			defaultCursor := fetcher.NewDefaultCursor(c, CreateGlucoseDatum)
			return fetcher.NewContinuousDeviceDataCursor(defaultCursor, dataFetcher, CreateGlucoseDatum)
		},
		dataFetcher: dataFetcher,
		summaries:   store.NewSummaries[*types.ContinuousStats, *types.ContinuousBucket](collection),
		buckets:     store.NewBuckets[*types.ContinuousBucket](bucketsCollection, types.SummaryTypeContinuous),
	}
}

func (gs *GlucoseSummarizer[A, P, T, B]) DeleteSummaries(ctx context.Context, userId string) error {
	return gs.summaries.DeleteSummary(ctx, userId)
}

func (gs *GlucoseSummarizer[A, P, T, B]) GetSummary(ctx context.Context, userId string) (*types.Summary[A, P, T, B], error) {
	return gs.summaries.GetSummary(ctx, userId)
}

func (gs *GlucoseSummarizer[A, P, T, B]) ClearInvalidatedBuckets(ctx context.Context, userId string, earliestModified time.Time) (time.Time, error) {
	return gs.buckets.ClearInvalidatedBuckets(ctx, userId, earliestModified)
}

func (gs *GlucoseSummarizer[A, P, T, B]) GetBucketsRange(ctx context.Context, userId string, startTime time.Time, endTime time.Time) (fetcher.AnyCursor, error) {
	return gs.buckets.GetBucketsRange(ctx, userId, &startTime, &endTime)
}

func (gs *GlucoseSummarizer[A, P, T, B]) SetOutdated(ctx context.Context, userId, reason string) (*time.Time, error) {
	return gs.summaries.SetOutdated(ctx, userId, reason)
}

func (gs *GlucoseSummarizer[A, P, T, B]) GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error) {
	return gs.summaries.GetOutdatedUserIDs(ctx, pagination)
}

func (gs *GlucoseSummarizer[A, P, T, B]) GetMigratableUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error) {
	return gs.summaries.GetMigratableUserIDs(ctx, pagination)
}

func (gs *GlucoseSummarizer[A, P, T, B]) BackfillSummaries(ctx context.Context) (int, error) {
	var empty struct{}

	distinctDataUserIDs, err := gs.dataFetcher.DistinctUserIDs(ctx, types.GetDeviceDataTypeStrings[A, P]())
	if err != nil {
		return 0, err
	}

	distinctSummaryIDs, err := gs.summaries.DistinctSummaryIDs(ctx)
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

	summaries := make([]*types.Summary[A, P, T, B], 0, len(userIDsReqBackfill))
	for _, userID := range userIDsReqBackfill {
		s := types.Create[A, P](userID)
		s.SetOutdated(types.OutdatedReasonBackfill)
		summaries = append(summaries, s)

		if len(summaries) >= backfillInsertBatch {
			count, err := gs.summaries.CreateSummaries(ctx, summaries)
			if err != nil {
				return count, err
			}
			summaries = nil
		}
	}

	if len(summaries) > 0 {
		return gs.summaries.CreateSummaries(ctx, summaries)
	}

	return 0, nil
}

func (gs *GlucoseSummarizer[A, P, T, B]) UpdateSummary(ctx context.Context, userId string) (*types.Summary[A, P, T, B], error) {
	logger := log.LoggerFromContext(ctx)
	userSummary, err := gs.GetSummary(ctx, userId)
	summaryType := types.GetDeviceDataTypeStrings[A, P]()
	if err != nil {
		return nil, err
	}

	logger.Debugf("Starting %s summary calculation for %s", types.GetTypeString[A, P](), userId)

	// user has no usable summary for incremental update
	if userSummary == nil {
		userSummary = types.Create[A, P](userId)
	}

	if userSummary.Stats == nil {
		userSummary.Stats = new(T)
		userSummary.Stats.Init()
	}

	if userSummary.Config.SchemaVersion != types.SchemaVersion {
		userSummary.SetOutdated(types.OutdatedReasonSchemaMigration)
		userSummary.Dates.Reset()
	}

	var status *data.UserDataStatus
	status, err = gs.dataFetcher.GetLastUpdatedForUser(ctx, userId, summaryType, userSummary.Dates.LastUpdatedDate)
	if err != nil {
		return nil, err
	}

	// this filters out users which cannot be updated, as they have no data of type T, but were called for update
	if status == nil {
		// user's data is inactive/ancient/deleted, or this summary shouldn't have been created
		logger.Warnf("User %s has a summary, but no data within range, deleting summary", userId)
		return nil, gs.summaries.DeleteSummary(ctx, userId)
	}

	// this filters out users which cannot be updated, as they somehow got called for update, but have no new data
	if status.EarliestModified.IsZero() {
		logger.Warnf("User %s was called for a %s summary update, but has no new data, skipping", userId, summaryType)

		userSummary.SetNotOutdated()
		return userSummary, gs.summaries.ReplaceSummary(ctx, userSummary)
	}

	if first, err := gs.buckets.ClearInvalidatedBuckets(ctx, userId, status.EarliestModified); err != nil {
		return nil, err
	} else if !first.IsZero() {
		status.FirstData = first
	}

	cursor, err := gs.dataFetcher.GetDataRange(ctx, userId, summaryType, status)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = userSummary.Stats.Update(ctx, userSummary.SummaryShared, gs.buckets, gs.cursorFactory(cursor))
	if err != nil {
		return nil, err
	}

	// this filters out users which may have appeared to have relevant data, but was filtered during calculation
	totalHours, err := gs.buckets.GetTotalHours(ctx, userId)
	if err != nil {
		return nil, err
	}
	if totalHours == 0 {
		logger.Warnf("User %s has a summary, but no valid data within range, creating placeholder summary", userId)
		userSummary.Dates.Reset()
		userSummary.Stats = nil
	} else {
		oldest, err := gs.buckets.GetOldestRecordTime(ctx, userId)
		if err != nil {
			return nil, err
		}
		userSummary.Dates.Update(status, oldest)
	}

	err = gs.summaries.ReplaceSummary(ctx, userSummary)

	return userSummary, err
}

func MaybeUpdateSummary(ctx context.Context, registry *SummarizerRegistry, updatesSummary map[string]struct{}, userId, reason string) map[string]*time.Time {
	outdatedSinceMap := make(map[string]*time.Time)
	lgr := log.LoggerFromContext(ctx)

	if _, ok := updatesSummary[types.SummaryTypeCGM]; ok {
		summarizer := GetSummarizer[*types.CGMStats, *types.GlucoseBucket](registry)
		outdatedSince, err := summarizer.SetOutdated(ctx, userId, reason)
		if err != nil {
			lgr.WithError(err).Error("Unable to set cgm summary outdated")
		}
		outdatedSinceMap[types.SummaryTypeCGM] = outdatedSince
	}

	if _, ok := updatesSummary[types.SummaryTypeBGM]; ok {
		summarizer := GetSummarizer[*types.BGMStats, *types.GlucoseBucket](registry)
		outdatedSince, err := summarizer.SetOutdated(ctx, userId, reason)
		if err != nil {
			lgr.WithError(err).Error("Unable to set bgm summary outdated")
		}
		outdatedSinceMap[types.SummaryTypeBGM] = outdatedSince
	}

	if _, ok := updatesSummary[types.SummaryTypeContinuous]; ok {
		summarizer := GetSummarizer[*types.ContinuousStats, *types.ContinuousBucket](registry)
		outdatedSince, err := summarizer.SetOutdated(ctx, userId, reason)
		if err != nil {
			lgr.WithError(err).Error("Unable to set continuous summary outdated")
		}
		outdatedSinceMap[types.SummaryTypeContinuous] = outdatedSince
	}

	return outdatedSinceMap
}

func CheckDatumUpdatesSummary(updatesSummary map[string]struct{}, datum data.Datum) {
	twoYearsPast := time.Now().UTC().AddDate(0, -24, 0)
	oneDayFuture := time.Now().UTC().AddDate(0, 0, 1)

	// we only update summaries if the data is both of a relevant type, and being uploaded as "active"
	// it also must be recent enough, within the past 2 years, and no more than 1d into the future
	if datum.IsActive() {
		typ := datum.GetType()
		if types.DeviceDataTypesSet.Contains(typ) && datum.GetTime().Before(oneDayFuture) && datum.GetTime().After(twoYearsPast) {
			updatesSummary[types.DeviceDataToSummaryTypes[typ]] = struct{}{}

			// Currently, both types update continuous summaries, this may need to be a separate check in the future.
			updatesSummary[types.SummaryTypeContinuous] = struct{}{}
		}
	}
}
