package v1

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tidepool-org/platform/summary"
	"github.com/tidepool-org/platform/summary/reporters"
	"github.com/tidepool-org/platform/summary/types"

	"github.com/tidepool-org/platform/clinics"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/api"
)

func SummaryRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.Get("/v1/summaries/cgm/:userId", GetSummary[*types.CGMPeriods, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/bgm/:userId", GetSummary[*types.BGMPeriods, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/continuous/:userId", GetSummary[*types.ContinuousPeriods, *types.ContinuousBucket], api.RequireAuth),

		dataService.Post("/v1/summaries/cgm/:userId", UpdateSummary[*types.CGMPeriods, *types.GlucoseBucket], api.RequireAuth),
		dataService.Post("/v1/summaries/bgm/:userId", UpdateSummary[*types.BGMPeriods, *types.GlucoseBucket], api.RequireAuth),
		dataService.Post("/v1/summaries/continuous/:userId", UpdateSummary[*types.ContinuousPeriods, *types.ContinuousBucket], api.RequireAuth),

		dataService.Get("/v1/summaries/outdated/cgm", GetOutdatedUserIDs[*types.CGMPeriods, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/outdated/bgm", GetOutdatedUserIDs[*types.BGMPeriods, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/outdated/continuous", GetOutdatedUserIDs[*types.ContinuousPeriods, *types.ContinuousBucket], api.RequireAuth),

		dataService.Get("/v1/summaries/migratable/cgm", GetMigratableUserIDs[*types.CGMPeriods, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/migratable/bgm", GetMigratableUserIDs[*types.BGMPeriods, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/migratable/continuous", GetMigratableUserIDs[*types.ContinuousPeriods, *types.ContinuousBucket], api.RequireAuth),

		dataService.Get("/v1/clinics/:clinicId/reports/realtime", GetPatientsWithRealtimeData, api.RequireAuth),
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

func GetSummary[PP types.PeriodsPt[P, PB, B], PB types.BucketDataPt[B], P types.Periods, B types.BucketData](dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	id := req.PathParam("userId")

	if !CheckPermissions(ctx, dataServiceContext, id) {
		return
	}

	summarizer := summary.GetSummarizer[PP, PB](dataServiceContext.SummarizerRegistry())
	userSummary, err := summarizer.GetSummary(ctx, id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
	} else if userSummary == nil {
		responder.Error(http.StatusNotFound, fmt.Errorf("no %s summary found for user %s", types.GetType[PP, PB](), id))
	} else {
		responder.Data(http.StatusOK, userSummary)
	}
}

func GetPatientsWithRealtimeData(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	clinicId := req.PathParam("clinicId")

	filter := reporters.NewPatientRealtimeDaysFilter()
	if err := request.DecodeRequestQuery(req.Request, filter); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	details := request.GetAuthDetails(ctx)

	if filter.StartTime.After(*filter.EndTime) {
		responder.Error(http.StatusBadRequest, errors.New("startTime is after endTime"))
		return
	}

	endOfHour := time.Now().Truncate(time.Hour).Add(time.Second * 3599)
	if filter.StartTime.Before(endOfHour.AddDate(0, 0, -60)) {
		responder.Error(http.StatusBadRequest, errors.New("startTime is too old ( >60d ago ) "))
		return
	}

	response, err := dataServiceContext.SummaryReporter().GetRealtimeDaysForPatients(
		ctx, dataServiceContext.ClinicsClient(), clinicId, details.Token(), *filter.StartTime, *filter.EndTime, filter.PatientFilters)
	if err != nil {
		if errors.Code(err) == clinics.ErrorCodeClinicClientFailure {
			res := errors.Meta(err).(*http.Response)
			responder.Reader(res.StatusCode, res.Body)
		} else {
			responder.Error(http.StatusInternalServerError, err)
		}
		return
	}

	responder.Data(http.StatusOK, response)
}

func UpdateSummary[PP types.PeriodsPt[P, PB, B], PB types.BucketDataPt[B], P types.Periods, B types.BucketData](dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	id := req.PathParam("userId")

	if !CheckPermissions(ctx, dataServiceContext, id) {
		return
	}

	summarizer := summary.GetSummarizer[PP, PB](dataServiceContext.SummarizerRegistry())
	userSummary, err := summarizer.UpdateSummary(ctx, id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
	} else {
		responder.Data(http.StatusOK, userSummary)
	}
}

func GetOutdatedUserIDs[PP types.PeriodsPt[P, PB, B], PB types.BucketDataPt[B], P types.Periods, B types.BucketData](dataServiceContext dataService.Context) {
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

	summarizer := summary.GetSummarizer[PP, PB](dataServiceContext.SummarizerRegistry())
	response, err := summarizer.GetOutdatedUserIDs(ctx, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, response)
}

func GetMigratableUserIDs[PP types.PeriodsPt[P, PB, B], PB types.BucketDataPt[B], P types.Periods, B types.BucketData](dataServiceContext dataService.Context) {
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

	summarizer := summary.GetSummarizer[PP, PB](dataServiceContext.SummarizerRegistry())
	userIDs, err := summarizer.GetMigratableUserIDs(ctx, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, userIDs)
}
