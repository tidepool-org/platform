package summary

import (
	"context"
	"github.com/tidepool-org/platform/errors"
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

func addSummarizer[PS types.ObservationsPt[S, PB, B], PB types.BucketDataPt[B], S types.Observations, B types.BucketData](reg *SummarizerRegistry, summarizer Summarizer[PS, PB, S, B]) {
	typ := types.GetType[PS, PB]()
	reg.summarizers[typ] = summarizer
}

func GetSummarizer[PS types.ObservationsPt[S, PB, B], PB types.BucketDataPt[B], S types.Observations, B types.BucketData](reg *SummarizerRegistry) Summarizer[PS, PB, S, B] {
	typ := types.GetType[PS, PB]()
	summarizer := reg.summarizers[typ]
	return summarizer.(Summarizer[PS, PB, S, B])
}

type Summarizer[PS types.ObservationsPt[S, PB, B], PB types.BucketDataPt[B], S types.Observations, B types.BucketData] interface {
	GetSummary(ctx context.Context, userId string) (*types.Summary[PS, PB, S, B], error)
	// TODO: Consider moving
	GetBucketsRange(ctx context.Context, userId string, startTime time.Time, endTime time.Time) (*mongo.Cursor, error)
	SetOutdated(ctx context.Context, userId, reason string) (*time.Time, error)
	UpdateSummary(ctx context.Context, userId string) (*types.Summary[PS, PB, S, B], error)
	GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error)
	GetMigratableUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error)
}

// Compile time interface check
var _ Summarizer[*types.CGMStats, *types.GlucoseBucket, types.CGMStats, types.GlucoseBucket] = &GlucoseSummarizer[*types.CGMStats, *types.GlucoseBucket, types.CGMStats, types.GlucoseBucket]{}
var _ Summarizer[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket] = &GlucoseSummarizer[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{}
var _ Summarizer[*types.ContinuousStats, *types.ContinuousBucket, types.ContinuousStats, types.ContinuousBucket] = &GlucoseSummarizer[*types.ContinuousStats, *types.ContinuousBucket, types.ContinuousStats, types.ContinuousBucket]{}

func CreateGlucoseDatum() data.Datum {
	return &glucoseDatum.Glucose{}
}

type GlucoseSummarizer[PS types.ObservationsPt[S, PB, B], PB types.BucketDataPt[B], S types.Observations, B types.BucketData] struct {
	cursorFactory fetcher.DataCursorFactory
	dataFetcher   fetcher.DeviceDataFetcher
	summaries     *store.Summaries[PS, PB, S, B]
	buckets       *store.Buckets[PB, B]
}

func NewBGMSummarizer(collection *storeStructuredMongo.Repository, bucketsCollection *storeStructuredMongo.Repository, dataFetcher fetcher.DeviceDataFetcher) Summarizer[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket] {
	return &GlucoseSummarizer[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
		cursorFactory: func(c *mongo.Cursor) fetcher.DeviceDataCursor {
			return fetcher.NewDefaultCursor(c, CreateGlucoseDatum)
		},
		dataFetcher: dataFetcher,
		summaries:   store.NewSummaries[*types.BGMPeriods, *types.GlucoseBucket](collection),
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

func (gs *GlucoseSummarizer[PS, PB, S, B]) DeleteSummaries(ctx context.Context, userId string) error {
	return gs.summaries.DeleteSummary(ctx, userId)
}

func (gs *GlucoseSummarizer[PS, PB, S, B]) GetSummary(ctx context.Context, userId string) (*types.Summary[PS, PB, S, B], error) {
	return gs.summaries.GetSummary(ctx, userId)
}

func (gs *GlucoseSummarizer[PS, PB, S, B]) GetBucketsRange(ctx context.Context, userId string, startTime time.Time, endTime time.Time) (*mongo.Cursor, error) {
	return gs.buckets.GetBucketsRange(ctx, userId, &startTime, &endTime)
}

func (gs *GlucoseSummarizer[PS, PB, S, B]) SetOutdated(ctx context.Context, userId, reason string) (*time.Time, error) {
	return gs.summaries.SetOutdated(ctx, userId, reason)
}

func (gs *GlucoseSummarizer[PS, PB, S, B]) GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error) {
	return gs.summaries.GetOutdatedUserIDs(ctx, pagination)
}

func (gs *GlucoseSummarizer[PS, PB, S, B]) GetMigratableUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error) {
	return gs.summaries.GetMigratableUserIDs(ctx, pagination)
}

func (gs *GlucoseSummarizer[PS, PB, S, B]) UpdateSummary(ctx context.Context, userId string) (*types.Summary[PS, PB, S, B], error) {
	logger := log.LoggerFromContext(ctx)
	userSummary, err := gs.GetSummary(ctx, userId)
	summaryType := types.GetType[PS, PB]()
	dataTypes := types.GetDeviceDataType[PS, PB]()
	if err != nil {
		return nil, err
	}

	logger.Debugf("Starting %s summary calculation for %s", types.GetType[PS, PB](), userId)

	// user has no usable summary for incremental update
	if userSummary == nil {
		userSummary = types.Create[PS, PB](userId)
	}

	if userSummary.Stats == nil {
		userSummary.Stats = new(S)
		userSummary.Stats.Init()
	}

	if userSummary.Config.SchemaVersion != types.SchemaVersion {
		userSummary.SetOutdated(types.OutdatedReasonSchemaMigration)
		userSummary.Dates.Reset()
	}

	var status *data.UserDataStatus
	status, err = gs.dataFetcher.GetLastUpdatedForUser(ctx, userId, dataTypes, userSummary.Dates.LastUpdatedDate)
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

	cursor, err := gs.dataFetcher.GetDataRange(ctx, userId, dataTypes, status)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = gs.UpdateBuckets(ctx, userId, summaryType, gs.cursorFactory(cursor))
	if err != nil {
		return nil, err
	}

	allBuckets, err := gs.buckets.GetAllBuckets(ctx, userId)
	if err != nil {
		return nil, err
	}

	err = userSummary.Stats.Update(ctx, allBuckets)
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

func (gs *GlucoseSummarizer[PS, PB, S, B]) UpdateBuckets(ctx context.Context, userId string, typ string, cursor fetcher.DeviceDataCursor) error {
	hasMoreData := true
	for hasMoreData {
		userData, err := cursor.GetNextBatch(ctx)
		if errors.Is(err, fetcher.ErrCursorExhausted) {
			hasMoreData = false
		} else if err != nil {
			return err
		}

		if len(userData) > 0 {
			startTime := userData[0].GetTime().UTC().Truncate(time.Hour)
			endTime := userData[len(userData)-1].GetTime().UTC().Truncate(time.Hour)
			buckets, err := gs.buckets.GetBucketsByTime(ctx, userId, startTime, endTime)
			if err != nil {
				return err
			}

			err = buckets.Update(types.CreateBucketForUser[PB](userId, typ), userData)
			if err != nil {
				return err
			}

			// TODO call this once? edge case of last bucket overlapping GetBucketsByTime result would have to be patched
			err = gs.buckets.WriteModifiedBuckets(ctx, buckets)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
		summarizer := GetSummarizer[*types.BGMPeriods, *types.GlucoseBucket](registry)
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
			for _, summaryType := range types.DeviceDataToSummaryTypes[typ] {
				updatesSummary[summaryType] = struct{}{}
			}
		}
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
