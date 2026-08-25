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
	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/api"
)

func SummaryRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.Get("/v1/summaries/cgm/:userId", GetSummary[*types.CGMPeriods, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/bgm/:userId", GetSummary[*types.BGMPeriods, *types.GlucoseBucket], api.RequireAuth),
		dataService.Get("/v1/summaries/con/:userId", GetSummary[*types.ContinuousPeriods, *types.ContinuousBucket], api.RequireAuth),

		dataService.Post("/v1/summaries/cgm/:userId", UpdateSummary, api.RequireAuth),
		dataService.Post("/v1/summaries/bgm/:userId", UpdateSummary, api.RequireAuth),
		dataService.Post("/v1/summaries/con/:userId", UpdateSummary, api.RequireAuth),

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
		responder.InternalServerError(err)
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
			responder.InternalServerError(err)
		}
		return
	}

	responder.Data(http.StatusOK, response)
}

// UpdateSummary reports the data of the user as changed rather than recalculating synchronously,
// which the retired task runners required. The work created recalculates every summary of the user.
func UpdateSummary(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	id := req.PathParam("userId")

	if !CheckPermissions(ctx, dataServiceContext, id) {
		return
	}

	if err := dataWorkPostprocess.Enqueue(ctx, dataServiceContext.WorkClient(), id, dataWorkPostprocess.ReasonDataAdded); err != nil {
		responder.InternalServerError(err)
		return
	}

	responder.Empty(http.StatusAccepted)
}
