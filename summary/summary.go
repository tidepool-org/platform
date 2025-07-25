package summary

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/summary/fetcher"
	"github.com/tidepool-org/platform/summary/store"
	"github.com/tidepool-org/platform/summary/types"
)

type SummarizerRegistry struct {
	summarizers map[string]any
}

func New(summaryRepository *storeStructuredMongo.Repository, bucketsRepository *storeStructuredMongo.Repository, fetcher fetcher.DeviceDataFetcher, mongoClient *mongo.Client) *SummarizerRegistry {
	registry := &SummarizerRegistry{summarizers: make(map[string]any)}
	addSummarizer(registry, NewCGMSummarizer(summaryRepository, bucketsRepository, fetcher, mongoClient))
	addSummarizer(registry, NewBGMSummarizer(summaryRepository, bucketsRepository, fetcher, mongoClient))
	addSummarizer(registry, NewContinuousSummarizer(summaryRepository, bucketsRepository, fetcher, mongoClient))
	return registry
}

func addSummarizer[PP types.PeriodsPt[P, PB, B], PB types.BucketDataPt[B], P types.Periods, B types.BucketData](reg *SummarizerRegistry, summarizer Summarizer[PP, PB, P, B]) {
	typ := types.GetType[PP, PB]()
	reg.summarizers[typ] = summarizer
}

func GetSummarizer[PP types.PeriodsPt[P, PB, B], PB types.BucketDataPt[B], P types.Periods, B types.BucketData](reg *SummarizerRegistry) Summarizer[PP, PB, P, B] {
	typ := types.GetType[PP, PB]()
	summarizer := reg.summarizers[typ]
	return summarizer.(Summarizer[PP, PB, P, B])
}

type Summarizer[PP types.PeriodsPt[P, PB, B], PB types.BucketDataPt[B], P types.Periods, B types.BucketData] interface {
	GetSummary(ctx context.Context, userId string) (*types.Summary[PP, PB, P, B], error)
	GetBucketsRange(ctx context.Context, userId string, startTime time.Time, endTime time.Time) (*mongo.Cursor, error)
	SetOutdated(ctx context.Context, userId, reason string) (*time.Time, error)
	UpdateSummary(ctx context.Context, userId string) (*types.Summary[PP, PB, P, B], error)
	GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error)
	GetMigratableUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error)
}

// Compile time interface check
var _ Summarizer[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket] = &GlucoseSummarizer[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{}
var _ Summarizer[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket] = &GlucoseSummarizer[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{}
var _ Summarizer[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket] = &GlucoseSummarizer[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{}

func CreateSelfMonitoredGlucoseDatum() data.Datum {
	return selfmonitored.New()
}

func CreateContinuousGlucoseDatum() data.Datum {
	return continuous.New()
}

type GlucoseSummarizer[PP types.PeriodsPt[P, PB, B], PB types.BucketDataPt[B], P types.Periods, B types.BucketData] struct {
	cursorFactory fetcher.DataCursorFactory
	dataFetcher   fetcher.DeviceDataFetcher
	summaries     *store.Summaries[PP, PB, P, B]
	buckets       *store.Buckets[PB, B]
	mongoClient   *mongo.Client
}

func NewBGMSummarizer(collection *storeStructuredMongo.Repository, bucketsCollection *storeStructuredMongo.Repository, dataFetcher fetcher.DeviceDataFetcher, mongoClient *mongo.Client) Summarizer[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket] {
	return &GlucoseSummarizer[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
		cursorFactory: func(c *mongo.Cursor) fetcher.DeviceDataCursor {
			return fetcher.NewDefaultCursor(c, CreateSelfMonitoredGlucoseDatum)
		},
		dataFetcher: dataFetcher,
		summaries:   store.NewSummaries[*types.BGMPeriods, *types.GlucoseBucket](collection),
		buckets:     store.NewBuckets[*types.GlucoseBucket](bucketsCollection, types.SummaryTypeBGM),
		mongoClient: mongoClient,
	}
}

func NewCGMSummarizer(collection *storeStructuredMongo.Repository, bucketsCollection *storeStructuredMongo.Repository, dataFetcher fetcher.DeviceDataFetcher, mongoClient *mongo.Client) Summarizer[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket] {
	return &GlucoseSummarizer[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
		cursorFactory: func(c *mongo.Cursor) fetcher.DeviceDataCursor {
			return fetcher.NewDefaultCursor(c, CreateContinuousGlucoseDatum)
		},
		dataFetcher: dataFetcher,
		summaries:   store.NewSummaries[*types.CGMPeriods, *types.GlucoseBucket](collection),
		buckets:     store.NewBuckets[*types.GlucoseBucket](bucketsCollection, types.SummaryTypeCGM),
		mongoClient: mongoClient,
	}
}

func (gs *GlucoseSummarizer[PP, PB, P, B]) DeleteSummaries(ctx context.Context, userId string) error {
	return gs.summaries.DeleteSummary(ctx, userId)
}

func (gs *GlucoseSummarizer[PP, PB, P, B]) GetSummary(ctx context.Context, userId string) (*types.Summary[PP, PB, P, B], error) {
	return gs.summaries.GetSummary(ctx, userId)
}

func (gs *GlucoseSummarizer[PP, PB, P, B]) GetBucketsRange(ctx context.Context, userId string, startTime time.Time, endTime time.Time) (*mongo.Cursor, error) {
	return gs.buckets.GetBucketsRange(ctx, userId, &startTime, &endTime)
}

func (gs *GlucoseSummarizer[PP, PB, P, B]) SetOutdated(ctx context.Context, userId, reason string) (*time.Time, error) {
	return gs.summaries.SetOutdated(ctx, userId, reason)
}

func (gs *GlucoseSummarizer[PP, PB, P, B]) GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error) {
	return gs.summaries.GetOutdatedUserIDs(ctx, pagination)
}

func (gs *GlucoseSummarizer[PP, PB, P, B]) GetMigratableUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error) {
	return gs.summaries.GetMigratableUserIDs(ctx, pagination)
}

func (gs *GlucoseSummarizer[PP, PB, P, B]) UpdateSummary(ctx context.Context, userId string) (*types.Summary[PP, PB, P, B], error) {
	logger := log.LoggerFromContext(ctx)
	result, err := store.WithTransaction(ctx, gs.mongoClient, func(sessionCtx mongo.SessionContext) (interface{}, error) {
		userSummary, err := gs.GetSummary(sessionCtx, userId)
		summaryType := types.GetType[PP, PB]()
		dataTypes := types.GetDeviceDataType[PP, PB]()
		if err != nil {
			return nil, err
		}

		logger.Debugf("Starting %s summary calculation for %s", types.GetType[PP, PB](), userId)

		// user has no usable summary for incremental update
		if userSummary == nil {
			userSummary = types.Create[PP, PB](userId)

			// This should be/will be a No-Op, but in case the summary was deleted without deleting buckets
			err = gs.buckets.Reset(sessionCtx, userId)
			if err != nil {
				return nil, err
			}
		}

		if userSummary.Periods == nil {
			userSummary.Periods = new(P)
			userSummary.Periods.Init()
		}

		if userSummary.Config.SchemaVersion != types.SchemaVersion {
			userSummary.SetOutdated(types.OutdatedReasonSchemaMigration)
			userSummary.Dates.Reset()

			// Drop all buckets for this user for a full reset
			err = gs.buckets.Reset(sessionCtx, userId)
			if err != nil {
				return nil, err
			}
		}

		var status *data.UserDataStatus
		status, err = gs.dataFetcher.GetLastUpdatedForUser(sessionCtx, userId, dataTypes, userSummary.Dates.LastUpdatedDate)
		if err != nil {
			return nil, err
		}

		// this filters out users which cannot be updated, as they have no data of type T, but were called for update
		if status == nil {
			// user's data is inactive/ancient/deleted, or this summary shouldn't have been created
			logger.Warnf("User %s has a summary, but no data within range, deleting summary", userId)
			return nil, gs.summaries.DeleteSummary(sessionCtx, userId)
		}

		// this filters out users which cannot be updated, as they somehow got called for update, but have no new data
		if status.EarliestModified.IsZero() {
			logger.Warnf("User %s was called for a %s summary update, but has no new data, skipping", userId, summaryType)

			userSummary.SetNotOutdated()
			return userSummary, gs.summaries.ReplaceSummary(sessionCtx, userSummary)
		}

		// only attempt to invalidate buckets if there is buckets which exist in the modified range
		if status.EarliestModified.Compare(userSummary.Dates.LastData) <= 0 {
			if newFirstData, err := gs.buckets.ClearInvalidatedBuckets(sessionCtx, userId, status.EarliestModified); err != nil {
				return nil, err
			} else if !newFirstData.IsZero() {
				status.FirstData = newFirstData
			}
		} else if userSummary.Dates.LastData.After(status.FirstData) {
			// otherwise limit FirstData to previous LastData
			status.FirstData = userSummary.Dates.LastData
		}

		cursor, err := gs.dataFetcher.GetDataRange(sessionCtx, userId, dataTypes, status)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(sessionCtx)

		err = gs.UpdateBuckets(sessionCtx, userId, summaryType, gs.cursorFactory(cursor))
		if err != nil {
			return nil, err
		}

		err = gs.buckets.TrimExcessBuckets(sessionCtx, userId)
		if err != nil {
			return nil, err
		}

		allBuckets, err := gs.buckets.GetAllBuckets(sessionCtx, userId)
		if err != nil {
			return nil, err
		}

		err = userSummary.Periods.Update(sessionCtx, allBuckets)
		if err != nil {
			return nil, err
		}

		// this filters out users which may have appeared to have relevant data, but was filtered during calculation
		totalHours, err := gs.buckets.GetTotalHours(sessionCtx, userId)
		if err != nil {
			return nil, err
		}

		if totalHours == 0 {
			logger.Warnf("User %s has a summary, but no valid data within range, creating placeholder summary", userId)
			userSummary.Dates.Reset()
			userSummary.Periods = nil
		} else {
			oldest, err := gs.buckets.GetOldestRecordTime(sessionCtx, userId)
			if err != nil {
				return nil, err
			}
			userSummary.Dates.Update(status, oldest)
		}

		return userSummary, gs.summaries.ReplaceSummary(sessionCtx, userSummary)
	})
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, err
	}

	return result.(*types.Summary[PP, PB, P, B]), err
}

func (gs *GlucoseSummarizer[PP, PB, P, B]) UpdateBuckets(ctx context.Context, userId string, summaryType string, cursor fetcher.DeviceDataCursor) error {
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

			err = buckets.Update(types.CreateBucketForUser[PB](userId, summaryType), userData)
			if err != nil {
				return err
			}

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
		summarizer := GetSummarizer[*types.CGMPeriods, *types.GlucoseBucket](registry)
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
		summarizer := GetSummarizer[*types.ContinuousPeriods, *types.ContinuousBucket](registry)
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

func NewContinuousSummarizer(collection *storeStructuredMongo.Repository, bucketsCollection *storeStructuredMongo.Repository, dataFetcher fetcher.DeviceDataFetcher, mongoClient *mongo.Client) Summarizer[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket] {
	return &GlucoseSummarizer[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
		cursorFactory: func(c *mongo.Cursor) fetcher.DeviceDataCursor {
			defaultCursor := fetcher.NewDefaultCursor(c, CreateContinuousGlucoseDatum)
			return fetcher.NewContinuousDeviceDataCursor(defaultCursor, dataFetcher, CreateContinuousGlucoseDatum)
		},
		dataFetcher: dataFetcher,
		summaries:   store.NewSummaries[*types.ContinuousPeriods, *types.ContinuousBucket](collection),
		buckets:     store.NewBuckets[*types.ContinuousBucket](bucketsCollection, types.SummaryTypeContinuous),
		mongoClient: mongoClient,
	}
}
