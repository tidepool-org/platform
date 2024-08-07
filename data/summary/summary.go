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

func New(summaryRepository *storeStructuredMongo.Repository, fetcher fetcher.DeviceDataFetcher) *SummarizerRegistry {
	registry := &SummarizerRegistry{summarizers: make(map[string]any)}
	addSummarizer(registry, NewBGMSummarizer(summaryRepository, fetcher))
	addSummarizer(registry, NewCGMSummarizer(summaryRepository, fetcher))
	addSummarizer(registry, NewContinuousSummarizer(summaryRepository, fetcher))
	return registry
}

func addSummarizer[A types.StatsPt[T], T types.Stats](reg *SummarizerRegistry, summarizer Summarizer[A, T]) {
	typ := types.GetTypeString[A, T]()
	reg.summarizers[typ] = summarizer
}

func GetSummarizer[A types.StatsPt[T], T types.Stats](reg *SummarizerRegistry) Summarizer[A, T] {
	typ := types.GetTypeString[A, T]()
	summarizer := reg.summarizers[typ]
	return summarizer.(Summarizer[A, T])
}

type Summarizer[A types.StatsPt[T], T types.Stats] interface {
	GetSummary(ctx context.Context, userId string) (*types.Summary[A, T], error)
	SetOutdated(ctx context.Context, userId, reason string) (*time.Time, error)
	UpdateSummary(ctx context.Context, userId string) (*types.Summary[A, T], error)
	GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error)
	GetMigratableUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error)
	BackfillSummaries(ctx context.Context) (int, error)
}

// Compile time interface check
var _ Summarizer[*types.CGMStats, types.CGMStats] = &GlucoseSummarizer[*types.CGMStats, types.CGMStats]{}
var _ Summarizer[*types.BGMStats, types.BGMStats] = &GlucoseSummarizer[*types.BGMStats, types.BGMStats]{}
var _ Summarizer[*types.ContinuousStats, types.ContinuousStats] = &GlucoseSummarizer[*types.ContinuousStats, types.ContinuousStats]{}

func CreateGlucoseDatum() data.Datum {
	return &glucoseDatum.Glucose{}
}

type GlucoseSummarizer[A types.StatsPt[T], T types.Stats] struct {
	cursorFactory fetcher.DataCursorFactory
	dataFetcher   fetcher.DeviceDataFetcher
	summaries     *store.Repo[A, T]
}

func NewBGMSummarizer(collection *storeStructuredMongo.Repository, dataFetcher fetcher.DeviceDataFetcher) Summarizer[*types.BGMStats, types.BGMStats] {
	return &GlucoseSummarizer[*types.BGMStats, types.BGMStats]{
		cursorFactory: func(c *mongo.Cursor) fetcher.DeviceDataCursor {
			return fetcher.NewDefaultCursor(c, CreateGlucoseDatum)
		},
		dataFetcher: dataFetcher,
		summaries:   store.New[*types.BGMStats](collection),
	}
}

func NewCGMSummarizer(collection *storeStructuredMongo.Repository, dataFetcher fetcher.DeviceDataFetcher) Summarizer[*types.CGMStats, types.CGMStats] {
	return &GlucoseSummarizer[*types.CGMStats, types.CGMStats]{
		cursorFactory: func(c *mongo.Cursor) fetcher.DeviceDataCursor {
			return fetcher.NewDefaultCursor(c, CreateGlucoseDatum)
		},
		dataFetcher: dataFetcher,
		summaries:   store.New[*types.CGMStats](collection),
	}
}

func NewContinuousSummarizer(collection *storeStructuredMongo.Repository, dataFetcher fetcher.DeviceDataFetcher) Summarizer[*types.ContinuousStats, types.ContinuousStats] {
	return &GlucoseSummarizer[*types.ContinuousStats, types.ContinuousStats]{
		cursorFactory: func(c *mongo.Cursor) fetcher.DeviceDataCursor {
			defaultCursor := fetcher.NewDefaultCursor(c, CreateGlucoseDatum)
			return fetcher.NewContinuousDeviceDataCursor(defaultCursor, dataFetcher, CreateGlucoseDatum)
		},
		dataFetcher: dataFetcher,
		summaries:   store.New[*types.ContinuousStats](collection),
	}
}

func (gs *GlucoseSummarizer[A, T]) DeleteSummaries(ctx context.Context, userId string) error {
	return gs.summaries.DeleteSummary(ctx, userId)
}

func (gs *GlucoseSummarizer[A, T]) GetSummary(ctx context.Context, userId string) (*types.Summary[A, T], error) {
	return gs.summaries.GetSummary(ctx, userId)
}

func (gs *GlucoseSummarizer[A, T]) SetOutdated(ctx context.Context, userId, reason string) (*time.Time, error) {
	return gs.summaries.SetOutdated(ctx, userId, reason)
}

func (gs *GlucoseSummarizer[A, T]) GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error) {
	return gs.summaries.GetOutdatedUserIDs(ctx, pagination)
}

func (gs *GlucoseSummarizer[A, T]) GetMigratableUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error) {
	return gs.summaries.GetMigratableUserIDs(ctx, pagination)
}

func (gs *GlucoseSummarizer[A, T]) BackfillSummaries(ctx context.Context) (int, error) {
	var empty struct{}

	distinctDataUserIDs, err := gs.dataFetcher.DistinctUserIDs(ctx, types.GetDeviceDataTypeStrings[A]())
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

	summaries := make([]*types.Summary[A, T], 0, len(userIDsReqBackfill))
	for _, userID := range userIDsReqBackfill {
		s := types.Create[A](userID)
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

func (gs *GlucoseSummarizer[A, T]) UpdateSummary(ctx context.Context, userId string) (*types.Summary[A, T], error) {
	logger := log.LoggerFromContext(ctx)
	userSummary, err := gs.GetSummary(ctx, userId)
	summaryType := types.GetDeviceDataTypeStrings[A]()
	if err != nil {
		return nil, err
	}

	logger.Debugf("Starting %s summary calculation for %s", types.GetTypeString[A](), userId)

	// user has no usable summary for incremental update
	if userSummary == nil {
		userSummary = types.Create[A](userId)
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

	if first := userSummary.Stats.ClearInvalidatedBuckets(status.EarliestModified); !first.IsZero() {
		status.FirstData = first
	}

	cursor, err := gs.dataFetcher.GetDataRange(ctx, userId, summaryType, status)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = userSummary.Stats.Update(ctx, gs.cursorFactory(cursor))
	if err != nil {
		return nil, err
	}

	// this filters out users which may have appeared to have relevant data, but was filtered during calculation
	if userSummary.Stats.GetBucketsLen() == 0 {
		logger.Warnf("User %s has a summary, but no valid data within range, creating placeholder summary", userId)
		userSummary.Dates.Reset()
		userSummary.Stats = nil
	} else {
		userSummary.Dates.Update(status, userSummary.Stats.GetBucketDate(0))
	}

	err = gs.summaries.ReplaceSummary(ctx, userSummary)

	return userSummary, err
}

func MaybeUpdateSummary(ctx context.Context, registry *SummarizerRegistry, updatesSummary map[string]struct{}, userId, reason string) map[string]*time.Time {
	outdatedSinceMap := make(map[string]*time.Time)
	lgr := log.LoggerFromContext(ctx)

	if _, ok := updatesSummary[types.SummaryTypeCGM]; ok {
		summarizer := GetSummarizer[*types.CGMStats](registry)
		outdatedSince, err := summarizer.SetOutdated(ctx, userId, reason)
		if err != nil {
			lgr.WithError(err).Error("Unable to set cgm summary outdated")
		}
		outdatedSinceMap[types.SummaryTypeCGM] = outdatedSince
	}

	if _, ok := updatesSummary[types.SummaryTypeBGM]; ok {
		summarizer := GetSummarizer[*types.BGMStats](registry)
		outdatedSince, err := summarizer.SetOutdated(ctx, userId, reason)
		if err != nil {
			lgr.WithError(err).Error("Unable to set bgm summary outdated")
		}
		outdatedSinceMap[types.SummaryTypeBGM] = outdatedSince
	}

	if _, ok := updatesSummary[types.SummaryTypeContinuous]; ok {
		summarizer := GetSummarizer[*types.ContinuousStats](registry)
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
