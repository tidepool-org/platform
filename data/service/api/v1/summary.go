package v1

import (
	"context"
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/summary"
	"github.com/tidepool-org/platform/data/summary/types"

	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

func SummaryRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.MakeRoute("GET", "/v1/summaries/cgm/:userId", Authenticate(GetSummary[types.CGMStats, *types.CGMStats])),
		dataService.MakeRoute("GET", "/v1/summaries/bgm/:userId", Authenticate(GetSummary[types.BGMStats, *types.BGMStats])),

		dataService.MakeRoute("POST", "/v1/summaries/cgm/:userId", Authenticate(UpdateSummary[types.CGMStats, *types.CGMStats])),
		dataService.MakeRoute("POST", "/v1/summaries/bgm/:userId", Authenticate(UpdateSummary[types.BGMStats, *types.BGMStats])),

		dataService.MakeRoute("POST", "/v1/summaries/cgm", Authenticate(BackfillSummaries[types.CGMStats, *types.CGMStats])),
		dataService.MakeRoute("POST", "/v1/summaries/bgm", Authenticate(BackfillSummaries[types.BGMStats, *types.BGMStats])),

		dataService.MakeRoute("GET", "/v1/summaries/cgm", Authenticate(GetOutdatedUserIDs[types.CGMStats, *types.CGMStats])),
		dataService.MakeRoute("GET", "/v1/summaries/bgm", Authenticate(GetOutdatedUserIDs[types.BGMStats, *types.BGMStats])),
	}
}

func CheckPermissions(ctx context.Context, dataServiceContext dataService.Context, id string) bool {
	details := request.DetailsFromContext(ctx)

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
	summary, err := summarizer.GetSummary(ctx, id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
	} else if summary == nil {
		responder.Empty(http.StatusNotFound)
	} else {
		responder.Data(http.StatusOK, summary)
	}
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
	summary, err := summarizer.UpdateSummary(ctx, id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
	} else if summary == nil {
		responder.Empty(http.StatusNotFound)
	} else {
		responder.Data(http.StatusOK, summary)
	}
}

func BackfillSummaries[T types.Stats, A types.StatsPt[T]](dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	if details := request.DetailsFromContext(ctx); !details.IsService() {
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

	if details := request.DetailsFromContext(ctx); !details.IsService() {
		dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		return
	}

	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	summarizer := summary.GetSummarizer[T, A](dataServiceContext.SummarizerRegistry())
	userIDs, err := summarizer.GetOutdatedUserIDs(ctx, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, userIDs)
}
