package summary

import (
	"context"
	"errors"
	"fmt"
	clinic "github.com/tidepool-org/clinic/client"
	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/pointer"
	"time"

	"github.com/tidepool-org/platform/data"

	"github.com/tidepool-org/platform/data/summary/store"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

const (
	backfillBatch                 = 50000
	backfillInsertBatch           = 10000
	RealtimeUserThreshold         = 16
	realtimePatientsLengthLimit   = 1000
	realtimePatientsInsuranceCode = "CPT-99454"
)

type SummarizerRegistry struct {
	summarizers map[string]any
}

func New(summaryRepository *storeStructuredMongo.Repository, dataRepository types.DeviceDataFetcher) *SummarizerRegistry {
	registry := &SummarizerRegistry{summarizers: make(map[string]any)}
	addSummarizer(registry, NewBGMSummarizer(summaryRepository, dataRepository))
	addSummarizer(registry, NewCGMSummarizer(summaryRepository, dataRepository))
	addSummarizer(registry, NewContinuousSummarizer(summaryRepository, dataRepository))
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

type Reporter struct {
	summarizer Summarizer[*types.ContinuousStats, types.ContinuousStats]
}

func NewReporter(registry *SummarizerRegistry) *Reporter {
	summarizer := GetSummarizer[*types.ContinuousStats](registry)
	return &Reporter{
		summarizer: summarizer,
	}
}

func (r *Reporter) GetRealtimeDaysForPatients(ctx context.Context, clinicsClient clinics.Client, clinicId string, token string, startTime time.Time, endTime time.Time) (*RealtimePatientsResponse, error) {
	params := &clinic.ListPatientsParams{
		Limit: pointer.FromAny(realtimePatientsLengthLimit + 1),
	}

	patients, err := clinicsClient.GetPatients(ctx, clinicId, token, params)
	if err != nil {
		return nil, err
	}

	if len(patients) > realtimePatientsLengthLimit {
		return nil, fmt.Errorf("too many patients in clinic for report to succeed. (%d > limit %d)", len(patients), realtimePatientsLengthLimit)
	}

	userIds := make([]string, len(patients))
	for i := 0; i < len(patients); i++ {
		userIds[i] = *patients[0].Id
	}

	userIdsRealtimeDays, err := r.GetRealtimeDaysForUsers(ctx, userIds, startTime, endTime)
	if err != nil {
		return nil, err
	}

	patientsResponse := make([]RealtimePatientResponse, len(userIdsRealtimeDays))
	for i := 0; i < len(userIdsRealtimeDays); i++ {
		patientsResponse[i] = RealtimePatientResponse{
			Id:                *patients[i].Id,
			FullName:          patients[i].FullName,
			BirthDate:         patients[i].BirthDate.Time,
			MRN:               patients[i].Mrn,
			RealtimeDays:      userIdsRealtimeDays[*patients[i].Id],
			HasSufficientData: userIdsRealtimeDays[*patients[i].Id] >= RealtimeUserThreshold,
		}
	}

	return &RealtimePatientsResponse{
		Config: RealtimePatientConfigResponse{
			Code:      realtimePatientsInsuranceCode,
			ClinicId:  clinicId,
			StartDate: startTime,
			EndDate:   endTime,
		},
		Results: patientsResponse,
	}, nil
}

func (r *Reporter) GetRealtimeDaysForUsers(ctx context.Context, userIds []string, startTime time.Time, endTime time.Time) (map[string]int, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userIds == nil {
		return nil, errors.New("userIds is missing")
	}
	if len(userIds) == 0 {
		return nil, errors.New("no userIds provided")
	}
	if startTime.IsZero() {
		return nil, errors.New("startTime is missing")
	}
	if endTime.IsZero() {
		return nil, errors.New("startTime is missing")
	}

	if startTime.After(endTime) {
		return nil, errors.New("startTime is after endTime")
	}

	if startTime.Before(time.Now().AddDate(0, 0, -60)) {
		return nil, errors.New("startTime is too old ( >60d ago ) ")
	}

	if int(endTime.Sub(startTime).Hours()/24) < RealtimeUserThreshold {
		return nil, errors.New("time range smaller than threshold, impossible")
	}

	realtimeUsers := make(map[string]int)

	for _, userId := range userIds {
		userSummary, err := r.summarizer.GetSummary(ctx, userId)
		if err != nil {
			return nil, err
		}

		realtimeUsers[userId] = userSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)

	}

	return realtimeUsers, nil
}

// Compile time interface check
var _ Summarizer[*types.CGMStats, types.CGMStats] = &GlucoseSummarizer[*types.CGMStats, types.CGMStats]{}
var _ Summarizer[*types.BGMStats, types.BGMStats] = &GlucoseSummarizer[*types.BGMStats, types.BGMStats]{}
var _ Summarizer[*types.ContinuousStats, types.ContinuousStats] = &GlucoseSummarizer[*types.ContinuousStats, types.ContinuousStats]{}

type GlucoseSummarizer[A types.StatsPt[T], T types.Stats] struct {
	userData  types.DeviceDataFetcher
	summaries *store.Repo[A, T]
}

func NewBGMSummarizer(collection *storeStructuredMongo.Repository, dataRepo types.DeviceDataFetcher) Summarizer[*types.BGMStats, types.BGMStats] {
	return &GlucoseSummarizer[*types.BGMStats, types.BGMStats]{
		userData:  dataRepo,
		summaries: store.New[*types.BGMStats](collection),
	}
}

func NewCGMSummarizer(collection *storeStructuredMongo.Repository, dataRepo types.DeviceDataFetcher) Summarizer[*types.CGMStats, types.CGMStats] {
	return &GlucoseSummarizer[*types.CGMStats, types.CGMStats]{
		userData:  dataRepo,
		summaries: store.New[*types.CGMStats](collection),
	}
}

func NewContinuousSummarizer(collection *storeStructuredMongo.Repository, dataRepo types.DeviceDataFetcher) Summarizer[*types.ContinuousStats, types.ContinuousStats] {
	return &GlucoseSummarizer[*types.ContinuousStats, types.ContinuousStats]{
		userData:  dataRepo,
		summaries: store.New[*types.ContinuousStats](collection),
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

	distinctDataUserIDs, err := gs.userData.DistinctUserIDs(ctx, types.GetDeviceDataTypeStrings[A]())
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

	if userSummary.Config.SchemaVersion != types.SchemaVersion {
		userSummary.SetOutdated(types.OutdatedReasonSchemaMigration)
		userSummary.Dates.Reset()
	}

	var status *data.UserDataStatus
	status, err = gs.userData.GetLastUpdatedForUser(ctx, userId, summaryType, userSummary.Dates.LastUpdatedDate)
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

	var cursor types.DeviceDataCursor
	cursor, err = gs.userData.GetDataRange(ctx, userId, summaryType, status)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = userSummary.Stats.Update(ctx, cursor, gs.userData)
	if err != nil {
		return nil, err
	}

	// this filters out users which may have appeared to have relevant data, but was filtered during calculation
	if userSummary.Stats.GetBucketsLen() == 0 {
		logger.Warnf("User %s has a summary, but no valid data within range, deleting summary", userId)
		return nil, gs.summaries.DeleteSummary(ctx, userId)
	}

	userSummary.Dates.Update(status, userSummary.Stats.GetBucketDate(0))

	err = gs.summaries.ReplaceSummary(ctx, userSummary)

	return userSummary, err
}

func MaybeUpdateSummary(ctx context.Context, registry *SummarizerRegistry, updatesSummary map[string]struct{}, userId, reason string) map[string]*time.Time {
	outdatedSinceMap := make(map[string]*time.Time)
	lgr := log.LoggerFromContext(ctx)

	if _, ok := updatesSummary[types.SummaryTypeCGM]; ok {
		summarizer := GetSummarizer[*types.CGMStats, types.CGMStats](registry)
		outdatedSince, err := summarizer.SetOutdated(ctx, userId, reason)
		if err != nil {
			lgr.WithError(err).Error("Unable to set cgm summary outdated")
		}
		outdatedSinceMap[types.SummaryTypeCGM] = outdatedSince
	}

	if _, ok := updatesSummary[types.SummaryTypeBGM]; ok {
		summarizer := GetSummarizer[*types.BGMStats, types.BGMStats](registry)
		outdatedSince, err := summarizer.SetOutdated(ctx, userId, reason)
		if err != nil {
			lgr.WithError(err).Error("Unable to set bgm summary outdated")
		}
		outdatedSinceMap[types.SummaryTypeBGM] = outdatedSince
	}

	if _, ok := updatesSummary[types.SummaryTypeContinuous]; ok {
		summarizer := GetSummarizer[*types.ContinuousStats, types.ContinuousStats](registry)
		outdatedSince, err := summarizer.SetOutdated(ctx, userId, reason)
		if err != nil {
			lgr.WithError(err).Error("Unable to set bgm summary outdated")
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
		}
	}
}
