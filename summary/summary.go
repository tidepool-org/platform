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
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/summary/fetcher"
	"github.com/tidepool-org/platform/summary/store"
	"github.com/tidepool-org/platform/summary/types"
)

//go:generate mockgen -source=summary.go -destination=test/summary_mocks.go -package=test -typed

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
	GetSummary(ctx context.Context, userID string) (*types.Summary[PP, PB, P, B], error)
	GetBucketsRange(ctx context.Context, userID string, startTime time.Time, endTime time.Time) (*mongo.Cursor, error)
	UpdateSummary(ctx context.Context, userID string) (*types.Summary[PP, PB, P, B], error)
	// RecalculateSummary recalculates the summary of the user from all of their data,
	// rather than incrementally from the data modified since it was last updated.
	RecalculateSummary(ctx context.Context, userID string) (*types.Summary[PP, PB, P, B], error)
}

// Compile time interface check
var _ Summarizer[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket] = &GlucoseSummarizer[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{}
var _ Summarizer[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket] = &GlucoseSummarizer[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{}
var _ Summarizer[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket] = &GlucoseSummarizer[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{}

func CreateSelfMonitoredGlucoseDatum(typ string) (data.Datum, error) {
	return selfmonitored.New(), nil
}

func CreateContinuousGlucoseDatum(typ string) (data.Datum, error) {
	return continuous.New(), nil
}

func CreateGlucoseDatum(typ string) (data.Datum, error) {
	switch typ {
	case continuous.Type:
		return CreateContinuousGlucoseDatum(typ)
	case selfmonitored.Type:
		return CreateSelfMonitoredGlucoseDatum(typ)
	default:
		return nil, errors.Newf("unknown glucose datum type %s", typ)
	}
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

func (gs *GlucoseSummarizer[PP, PB, P, B]) DeleteSummaries(ctx context.Context, userID string) error {
	return gs.summaries.DeleteSummary(ctx, userID)
}

func (gs *GlucoseSummarizer[PP, PB, P, B]) GetSummary(ctx context.Context, userID string) (*types.Summary[PP, PB, P, B], error) {
	return gs.summaries.GetSummary(ctx, userID)
}

func (gs *GlucoseSummarizer[PP, PB, P, B]) GetBucketsRange(ctx context.Context, userID string, startTime time.Time, endTime time.Time) (*mongo.Cursor, error) {
	return gs.buckets.GetBucketsRange(ctx, userID, &startTime, &endTime)
}

func (gs *GlucoseSummarizer[PP, PB, P, B]) UpdateSummary(ctx context.Context, userID string) (*types.Summary[PP, PB, P, B], error) {
	return gs.updateSummary(ctx, userID, false)
}

func (gs *GlucoseSummarizer[PP, PB, P, B]) RecalculateSummary(ctx context.Context, userID string) (*types.Summary[PP, PB, P, B], error) {
	return gs.updateSummary(ctx, userID, true)
}

// updateSummary updates the summary of the user incrementally, unless recalculate is true
// or the summary's schema version is outdated, in which case it's recalculated from all of
// the user's data.
func (gs *GlucoseSummarizer[PP, PB, P, B]) updateSummary(ctx context.Context, userID string, recalculate bool) (*types.Summary[PP, PB, P, B], error) {
	logger := log.LoggerFromContext(ctx)
	result, err := storeStructuredMongo.WithTransaction(ctx, gs.mongoClient, func(sessionCtx mongo.SessionContext) (interface{}, error) {
		now := time.Now()
		summaryType := types.GetType[PP, PB]()
		dataTypes := types.GetDeviceDataType[PP, PB]()
		userSummary, reset, err := gs.getOrInitSummary(sessionCtx, userID, recalculate)
		if err != nil {
			return nil, err
		}

		logger.Debugf("Starting %s summary calculation for %s", summaryType, userID)

		if reset != resetNone {
			err = gs.buckets.Reset(sessionCtx, userID)
			if err != nil {
				return nil, err
			}
		}

		var status *data.UserDataStatus
		status, err = gs.dataFetcher.GetUserDataStatus(sessionCtx, userID, dataTypes, userSummary.Dates.LastUpdatedDate)
		if err != nil {
			return nil, err
		}

		// this filters out users which cannot be updated, as they have no data of type T, but were called for update
		if status == nil {
			// user's data is inactive/ancient/deleted, or this summary shouldn't have been created
			logger.Warnf("User %s has a summary, but no data within range, deleting summary", userID)
			return nil, gs.summaries.DeleteSummary(sessionCtx, userID)
		}

		// this filters out users which cannot be updated, as they somehow got called for update, but have no new data
		if status.EarliestModified.IsZero() {
			logger.Warnf("User %s was called for a %s summary update, but has no new data, skipping", userID, summaryType)
			return userSummary, nil
		}

		status.FirstData, err = gs.invalidateBuckets(sessionCtx, userID,
			userSummary.Dates.LastData, status)
		if err != nil {
			return nil, err
		}

		cursor, err := gs.dataFetcher.GetDataRange(sessionCtx, userID, dataTypes, status)
		if err != nil {
			return nil, err
		}
		defer storeStructuredMongo.CloseCursor(sessionCtx, cursor)

		err = gs.UpdateBuckets(sessionCtx, userID, summaryType, gs.cursorFactory(cursor))
		if err != nil {
			return nil, err
		}

		err = gs.buckets.TrimExcessBuckets(sessionCtx, userID)
		if err != nil {
			return nil, err
		}

		// All the data in the range may have been skipped when the buckets were updated, in
		// which case no bucket has data, and the summary is deleted.
		firstData, err := gs.buckets.GetFirstData(sessionCtx, userID)
		if err != nil {
			return nil, err
		}

		if firstData.IsZero() {
			logger.Warnf("User %s has a summary, but no valid data within range, deleting summary", userID)
			return nil, gs.summaries.DeleteSummary(sessionCtx, userID)
		}
		userSummary.Dates.Update(status, firstData)

		allBuckets, err := gs.buckets.GetAllBuckets(sessionCtx, userID)
		if err != nil {
			return nil, err
		}
		defer storeStructuredMongo.CloseCursor(sessionCtx, allBuckets)

		err = userSummary.Periods.Update(sessionCtx, allBuckets)
		if err != nil {
			return nil, err
		}

		if err := gs.summaries.ReplaceSummary(sessionCtx, userSummary); err != nil {
			return nil, err
		}

		logger.WithFields(log.Fields{
			"summaryType":     summaryType,
			"schemaMigration": reset == resetSchemaMigration,
			"recalculation":   recalculate,
			"duration":        time.Since(now).Microseconds(),
		}).Info("UpdateSummary")

		return userSummary, nil
	})
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, err
	}

	return result.(*types.Summary[PP, PB, P, B]), err
}

// resetReason is why the buckets of a user must be reset before their summary is calculated
type resetReason int

const (
	// resetNone keeps the buckets, as the stored summary is updated incrementally
	resetNone resetReason = iota
	// resetNewSummary drops any buckets left by a summary deleted without deleting them
	resetNewSummary
	// resetSchemaMigration drops the buckets of a summary calculated with an outdated
	// schema
	resetSchemaMigration
	// resetRecalculation drops the buckets of a summary whose recalculation was requested
	resetRecalculation
)

// getOrInitSummary returns the stored summary of the user, or a new summary when there's
// none or it must be recalculated from all of the user's data, along with why the user's
// buckets must be reset. A summary calculated with an outdated schema reports a schema
// migration even when its recalculation was also requested.
func (gs *GlucoseSummarizer[PP, PB, P, B]) getOrInitSummary(ctx context.Context,
	userID string, recalculate bool) (*types.Summary[PP, PB, P, B], resetReason, error) {

	userSummary, err := gs.GetSummary(ctx, userID)
	if err != nil {
		return nil, resetNone, err
	}

	switch {
	case userSummary == nil:
		return types.Create[PP, PB](userID), resetNewSummary, nil
	case userSummary.Config.SchemaVersion != types.SchemaVersion:
		return types.Create[PP, PB](userID), resetSchemaMigration, nil
	case recalculate:
		return types.Create[PP, PB](userID), resetRecalculation, nil
	}

	if userSummary.Periods == nil {
		userSummary.Periods = new(P)
		userSummary.Periods.Init()
	}

	return userSummary, resetNone, nil
}

// invalidateBuckets drops the buckets invalidated by data modified at or before lastData,
// the last data summarized, and returns the time after which the data must be read.
func (gs *GlucoseSummarizer[PP, PB, P, B]) invalidateBuckets(ctx context.Context,
	userID string, lastData time.Time, status *data.UserDataStatus) (time.Time, error) {

	if status.EarliestModified.After(lastData) {
		// Only data after the last data summarized was modified, so only the data after it
		// is read, unless the range starts later.
		if lastData.After(status.FirstData) {
			return lastData, nil
		}
		return status.FirstData, nil
	}

	// The modified data invalidates the buckets from its hour on. The data is read again
	// from the end of the newest bucket kept, or from the start of the range when no
	// bucket was dropped or none is kept.
	newFirstData, err := gs.buckets.ClearInvalidatedBuckets(ctx, userID,
		status.EarliestModified)
	if err != nil {
		return time.Time{}, err
	}
	if newFirstData.IsZero() {
		return status.FirstData, nil
	}
	return newFirstData, nil
}

func (gs *GlucoseSummarizer[PP, PB, P, B]) UpdateBuckets(ctx context.Context, userID string, summaryType string, cursor fetcher.DeviceDataCursor) error {
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
			buckets, err := gs.buckets.GetBucketsByTime(ctx, userID, startTime, endTime)
			if err != nil {
				return err
			}

			err = buckets.Update(types.CreateBucketForUser[PB](userID, summaryType), userData)
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

func NewContinuousSummarizer(collection *storeStructuredMongo.Repository, bucketsCollection *storeStructuredMongo.Repository, dataFetcher fetcher.DeviceDataFetcher, mongoClient *mongo.Client) Summarizer[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket] {
	return &GlucoseSummarizer[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
		cursorFactory: func(c *mongo.Cursor) fetcher.DeviceDataCursor {
			defaultCursor := fetcher.NewDefaultCursor(c, CreateGlucoseDatum)
			return fetcher.NewContinuousDeviceDataCursor(defaultCursor, dataFetcher, CreateGlucoseDatum)
		},
		dataFetcher: dataFetcher,
		summaries:   store.NewSummaries[*types.ContinuousPeriods, *types.ContinuousBucket](collection),
		buckets:     store.NewBuckets[*types.ContinuousBucket](bucketsCollection, types.SummaryTypeContinuous),
		mongoClient: mongoClient,
	}
}
