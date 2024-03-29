package summary

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/data"

	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/summary/store"
	"github.com/tidepool-org/platform/data/summary/types"
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
	SetOutdated(ctx context.Context, userId, reason string) (*time.Time, error)
	UpdateSummary(ctx context.Context, userId string) (*types.Summary[T, A], error)
	GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error)
	GetMigratableUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error)
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

func (c *GlucoseSummarizer[T, A]) SetOutdated(ctx context.Context, userId, reason string) (*time.Time, error) {
	return c.summaries.SetOutdated(ctx, userId, reason)
}

func (c *GlucoseSummarizer[T, A]) GetOutdatedUserIDs(ctx context.Context, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error) {
	return c.summaries.GetOutdatedUserIDs(ctx, pagination)
}

func (c *GlucoseSummarizer[T, A]) GetMigratableUserIDs(ctx context.Context, pagination *page.Pagination) ([]string, error) {
	return c.summaries.GetMigratableUserIDs(ctx, pagination)
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
		s := types.Create[A](userID)
		s.SetOutdated(types.OutdatedReasonBackfill)
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
		return nil, err
	}

	logger.Debugf("Starting %s summary calculation for %s", types.GetTypeString[T, A](), userId)

	// user has no usable summary for incremental update
	if userSummary == nil {
		userSummary = types.Create[A](userId)
	}

	if userSummary.Config.SchemaVersion != types.SchemaVersion {
		userSummary.SetOutdated(types.OutdatedReasonSchemaMigration)
		userSummary.Dates.Reset()
	}

	var status *types.UserLastUpdated
	status, err = c.deviceData.GetLastUpdatedForUser(ctx, userId, types.GetDeviceDataTypeString[T, A](), userSummary.Dates.LastUpdatedDate)
	if err != nil {
		return nil, err
	}

	// this filters out users which cannot be updated, as they have no data of type T, but were called for update
	if status.LastData.IsZero() {
		// user's data is inactive/ancient/deleted, or this summary shouldn't have been created
		logger.Warnf("User %s has a summary, but no data within range, deleting summary", userId)
		return nil, c.summaries.DeleteSummary(ctx, userId)
	}

	// this filters out users which cannot be updated, as they somehow got called for update, but have no new data
	if status.EarliestModified.IsZero() {
		logger.Warnf("User %s was called for a %s summary update, but has no new data, skipping", userId, types.GetTypeString[T, A]())
		userSummary.Dates.Update(status, userSummary.Stats.GetBucketDate(0))
		return userSummary, c.summaries.ReplaceSummary(ctx, userSummary)
	}

	if first := userSummary.Stats.ClearInvalidatedBuckets(status.EarliestModified); !first.IsZero() {
		status.FirstData = first
	}

	var cursor types.DeviceDataCursor
	cursor, err = c.deviceData.GetDataRange(ctx, userId, types.GetDeviceDataTypeString[T, A](), status)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = userSummary.Stats.Update(ctx, cursor)
	if err != nil {
		return nil, err
	}

	userSummary.Dates.Update(status, userSummary.Stats.GetBucketDate(0))

	err = c.summaries.ReplaceSummary(ctx, userSummary)

	return userSummary, err
}

func MaybeUpdateSummary(ctx context.Context, registry *SummarizerRegistry, updatesSummary map[string]struct{}, userId, reason string) map[string]*time.Time {
	outdatedSinceMap := make(map[string]*time.Time)
	lgr := log.LoggerFromContext(ctx)

	if _, ok := updatesSummary[types.SummaryTypeCGM]; ok {
		summarizer := GetSummarizer[types.CGMStats, *types.CGMStats](registry)
		outdatedSince, err := summarizer.SetOutdated(ctx, userId, reason)
		if err != nil {
			lgr.WithError(err).Error("Unable to set cgm summary outdated")
		}
		outdatedSinceMap[types.SummaryTypeCGM] = outdatedSince
	}

	if _, ok := updatesSummary[types.SummaryTypeBGM]; ok {
		summarizer := GetSummarizer[types.BGMStats, *types.BGMStats](registry)
		outdatedSince, err := summarizer.SetOutdated(ctx, userId, reason)
		if err != nil {
			lgr.WithError(err).Error("Unable to set bgm summary outdated")
		}
		outdatedSinceMap[types.SummaryTypeBGM] = outdatedSince
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
		}
	}
}

func CheckDataSetUpdatesSummary(ctx context.Context, repository dataStore.DataRepository, updatesSummary map[string]struct{}, dataSetId string) {
	twoYearsPast := time.Now().UTC().AddDate(0, -24, 0)
	oneDayFuture := time.Now().UTC().AddDate(0, 0, 1)

	for _, typ := range types.DeviceDataTypes {
		status, err := repository.CheckDataSetContainsTypeInRange(ctx, dataSetId, typ, twoYearsPast, oneDayFuture)
		if err != nil {
			return
		}
		if status {
			updatesSummary[types.DeviceDataToSummaryTypes[typ]] = struct{}{}
		}
	}
}
