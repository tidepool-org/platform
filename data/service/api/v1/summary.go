package v1

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tidepool-org/platform/clinics"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/summary"
	"github.com/tidepool-org/platform/data/summary/reporters"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/api"
)

func SummaryRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.Get("/v1/summaries/cgm/:userId", GetSummary[*types.CGMStats, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/bgm/:userId", GetSummary[*types.BGMStats, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/continuous/:userId", GetSummary[*types.ContinuousStats, *types.ContinuousBucket], api.RequireAuth),

		dataService.Post("/v1/summaries/cgm/:userId", UpdateSummary[*types.CGMStats, *types.GlucoseBucket], api.RequireAuth),
		dataService.Post("/v1/summaries/bgm/:userId", UpdateSummary[*types.BGMStats, *types.GlucoseBucket], api.RequireAuth),
		dataService.Post("/v1/summaries/continuous/:userId", UpdateSummary[*types.ContinuousStats, *types.ContinuousBucket], api.RequireAuth),

		dataService.Post("/v1/summaries/backfill/cgm", BackfillSummaries[*types.CGMStats, *types.GlucoseBucket], api.RequireAuth),
		dataService.Post("/v1/summaries/backfill/bgm", BackfillSummaries[*types.BGMStats, *types.GlucoseBucket], api.RequireAuth),
		dataService.Post("/v1/summaries/backfill/continuous", BackfillSummaries[*types.ContinuousStats, *types.ContinuousBucket], api.RequireAuth),

		dataService.Get("/v1/summaries/outdated/cgm", GetOutdatedUserIDs[*types.CGMStats, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/outdated/bgm", GetOutdatedUserIDs[*types.BGMStats, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/outdated/continuous", GetOutdatedUserIDs[*types.ContinuousStats, *types.ContinuousBucket], api.RequireAuth),

		dataService.Get("/v1/summaries/migratable/cgm", GetMigratableUserIDs[*types.CGMStats, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/migratable/bgm", GetMigratableUserIDs[*types.BGMStats, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/migratable/continuous", GetMigratableUserIDs[*types.ContinuousStats, *types.ContinuousBucket], api.RequireAuth),

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

func GetSummary[A types.StatsPt[T, P, B], P types.BucketDataPt[B], T types.Stats, B types.BucketData](dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	id := req.PathParam("userId")

	if !CheckPermissions(ctx, dataServiceContext, id) {
		return
	}

	summarizer := summary.GetSummarizer[A, P](dataServiceContext.SummarizerRegistry())
	userSummary, err := summarizer.GetSummary(ctx, id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
	} else if userSummary == nil {
		responder.Error(http.StatusNotFound, fmt.Errorf("no %s summary found for user %s", types.GetTypeString[A, P](), id))
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

func UpdateSummary[A types.StatsPt[T, P, B], P types.BucketDataPt[B], T types.Stats, B types.BucketData](dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	id := req.PathParam("userId")

	if !CheckPermissions(ctx, dataServiceContext, id) {
		return
	}

	summarizer := summary.GetSummarizer[A, P](dataServiceContext.SummarizerRegistry())
	userSummary, err := summarizer.UpdateSummary(ctx, id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
	} else {
		responder.Data(http.StatusOK, userSummary)
	}
}

func BackfillSummaries[A types.StatsPt[T, P, B], P types.BucketDataPt[B], T types.Stats, B types.BucketData](dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	if details := request.GetAuthDetails(ctx); !details.IsService() {
		dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		return
	}

	summarizer := summary.GetSummarizer[A, P](dataServiceContext.SummarizerRegistry())
	status, err := summarizer.BackfillSummaries(ctx)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, status)
}

func GetOutdatedUserIDs[A types.StatsPt[T, P, B], P types.BucketDataPt[B], T types.Stats, B types.BucketData](dataServiceContext dataService.Context) {
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

	summarizer := summary.GetSummarizer[A, P](dataServiceContext.SummarizerRegistry())
	response, err := summarizer.GetOutdatedUserIDs(ctx, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, response)
}

func GetMigratableUserIDs[A types.StatsPt[T, P, B], P types.BucketDataPt[B], T types.Stats, B types.BucketData](dataServiceContext dataService.Context) {
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

	summarizer := summary.GetSummarizer[A, P](dataServiceContext.SummarizerRegistry())
	userIDs, err := summarizer.GetMigratableUserIDs(ctx, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, userIDs)
}
