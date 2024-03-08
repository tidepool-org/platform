package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/tidepool-org/platform/data/summary/store"
	"github.com/tidepool-org/platform/structure"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/summary"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/service/api"

	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

const realtimePatientsInsuranceCode = "CPT-99454"

func SummaryRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.Get("/v1/summaries/cgm/:userId", GetSummary[types.CGMStats, *types.CGMStats], api.RequireAuth),
		dataService.Get("/v1/summaries/bgm/:userId", GetSummary[types.BGMStats, *types.BGMStats], api.RequireAuth),

		dataService.Post("/v1/summaries/cgm/:userId", UpdateSummary[types.CGMStats, *types.CGMStats], api.RequireAuth),
		dataService.Post("/v1/summaries/bgm/:userId", UpdateSummary[types.BGMStats, *types.BGMStats], api.RequireAuth),

		dataService.Post("/v1/summaries/backfill/cgm", BackfillSummaries[types.CGMStats, *types.CGMStats], api.RequireAuth),
		dataService.Post("/v1/summaries/backfill/bgm", BackfillSummaries[types.BGMStats, *types.BGMStats], api.RequireAuth),

		dataService.Get("/v1/summaries/outdated/cgm", GetOutdatedUserIDs[types.CGMStats, *types.CGMStats], api.RequireAuth),
		dataService.Get("/v1/summaries/outdated/bgm", GetOutdatedUserIDs[types.BGMStats, *types.BGMStats], api.RequireAuth),

		dataService.Get("/v1/summaries/migratable/cgm", GetMigratableUserIDs[types.CGMStats, *types.CGMStats], api.RequireAuth),
		dataService.Get("/v1/summaries/migratable/bgm", GetMigratableUserIDs[types.BGMStats, *types.BGMStats], api.RequireAuth),

		dataService.Get("/v1/summaries/realtime/:clinicId", GetPatientsWithRealtimeData, api.RequireAuth),
	}
}

func CheckPermissions(ctx context.Context, dataServiceContext dataService.Context, id string) bool {
	details := request.GetAuthDetails(ctx)

	if !details.IsService() {
		permissions, err := dataServiceContext.PermissionClient().GetUserPermissions(ctx, details.UserID(), id)
		if err != nil {
			if request.IsErrorUnauthorized(err) {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return false
		}
		if _, ok := permissions[permission.Read]; !ok {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return false
		}
	}
	return true
}

func GetSummary[T types.Stats, A types.StatsPt[T]](dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	id := req.PathParam("userId")

	if !CheckPermissions(ctx, dataServiceContext, id) {
		return
	}

	summarizer := summary.GetSummarizer[T, A](dataServiceContext.SummarizerRegistry())
	userSummary, err := summarizer.GetSummary(ctx, id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
	} else if userSummary == nil {
		responder.Empty(http.StatusNotFound)
	} else {
		responder.Data(http.StatusOK, userSummary)
	}
}

type RealtimePatientsResponse struct {
	Config  RealtimePatientConfigResponse `json:"config"`
	Results []RealtimePatientResponse     `json:"results"`
}

type RealtimePatientConfigResponse struct {
	Code      string    `json:"code"`
	ClinicId  string    `json:"clinicId"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type RealtimePatientResponse struct {
	Id                string    `json:"id"`
	FullName          string    `json:"fullName"`
	BirthDate         time.Time `json:"birthDate"`
	MRN               *string   `json:"mrn"`
	RealtimeDays      int       `json:"realtimeDays"`
	HasSufficientData bool      `json:"hasSufficientData"`
}

type RealtimePatientsFilter struct {
	StartTime *time.Time
	EndTime   *time.Time
}

func NewRealtimePatientsFilter() *RealtimePatientsFilter {
	return &RealtimePatientsFilter{}
}

func (d *RealtimePatientsFilter) Parse(parser structure.ObjectParser) {
	d.StartTime = parser.Time("startDate", time.RFC3339)
	d.EndTime = parser.Time("endDate", time.RFC3339)
}

func (d *RealtimePatientsFilter) Validate(validator structure.Validator) {
	validator.Time("startDate", d.StartTime).NotZero()
	validator.Time("endDate", d.EndTime).NotZero()
}

func GetPatientsWithRealtimeData(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	clinicId := req.PathParam("clinicId")

	filter := NewRealtimePatientsFilter()
	if err := request.DecodeRequestQuery(req.Request, filter); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	startTime := time.Now().UTC().AddDate(0, 0, -60)
	endTime := time.Now().UTC()

	details := request.GetAuthDetails(ctx)
	if !details.IsService() {
		dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		return
	}

	patients, err := dataServiceContext.ClinicsClient().GetPatients(ctx, clinicId, details.Token())
	userIds := make([]string, len(patients))
	for i := 0; i < len(patients); i++ {
		userIds[i] = *patients[0].Id
	}

	summaryManager := dataServiceContext.SummarizerRegistry().TypelessSummarizer
	userIdsRealtimeDays, err := summaryManager.GetPatientsWithRealtimeData(ctx, userIds, startTime, endTime)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	patientsResponse := make([]RealtimePatientResponse, len(patients))
	for i := 0; i < len(patients); i++ {
		patientsResponse[i] = RealtimePatientResponse{
			Id:                *patients[i].Id,
			FullName:          patients[i].FullName,
			BirthDate:         patients[i].BirthDate.Time,
			MRN:               patients[i].Mrn,
			RealtimeDays:      userIdsRealtimeDays[*patients[i].Id],
			HasSufficientData: userIdsRealtimeDays[*patients[i].Id] >= store.RealtimeUserThreshold,
		}
	}

	response := RealtimePatientsResponse{
		Config: RealtimePatientConfigResponse{
			Code:      realtimePatientsInsuranceCode,
			ClinicId:  clinicId,
			StartDate: startTime,
			EndDate:   endTime,
		},
		Results: patientsResponse,
	}

	responder.Data(http.StatusOK, response)
}

func UpdateSummary[T types.Stats, A types.StatsPt[T]](dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	id := req.PathParam("userId")

	if !CheckPermissions(ctx, dataServiceContext, id) {
		return
	}

	summarizer := summary.GetSummarizer[T, A](dataServiceContext.SummarizerRegistry())
	userSummary, err := summarizer.UpdateSummary(ctx, id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
	} else if userSummary == nil {
		responder.Empty(http.StatusNotFound)
	} else {
		responder.Data(http.StatusOK, userSummary)
	}
}

func BackfillSummaries[T types.Stats, A types.StatsPt[T]](dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	if details := request.GetAuthDetails(ctx); !details.IsService() {
		dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		return
	}

	summarizer := summary.GetSummarizer[T, A](dataServiceContext.SummarizerRegistry())
	status, err := summarizer.BackfillSummaries(ctx)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, status)
}

func GetOutdatedUserIDs[T types.Stats, A types.StatsPt[T]](dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	if details := request.GetAuthDetails(ctx); !details.IsService() {
		dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		return
	}

	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	summarizer := summary.GetSummarizer[T, A](dataServiceContext.SummarizerRegistry())
	response, err := summarizer.GetOutdatedUserIDs(ctx, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, response)
}

func GetMigratableUserIDs[T types.Stats, A types.StatsPt[T]](dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	if details := request.GetAuthDetails(ctx); !details.IsService() {
		dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		return
	}

	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	summarizer := summary.GetSummarizer[T, A](dataServiceContext.SummarizerRegistry())
	userIDs, err := summarizer.GetMigratableUserIDs(ctx, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, userIDs)
}
