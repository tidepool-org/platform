package v1

import (
	"net/http"
    "strconv"

	//"github.com/tidepool-org/platform/data"
	dataService "github.com/tidepool-org/platform/data/service"
	//"github.com/tidepool-org/platform/page"
	//"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
)

func SummaryRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.MakeRoute("GET", "/v1/summary/:userId", Authenticate(GetSummary)),
		dataService.MakeRoute("POST", "/v1/summary/:userId", Authenticate(UpdateSummary)),
		dataService.MakeRoute("GET", "/v1/agedsummaries/:minutes", Authenticate(GetAgedSummaries)),
	}
}

func GetSummary(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	dataClient := dataServiceContext.DataClient()

	//details := request.DetailsFromContext(req.Context())
	/*if details == nil {
		request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
		return
	}*/

	responder := request.MustNewResponder(res, req)

	id := req.PathParam("userId")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	summary, err := dataClient.GetSummary(req.Context(), id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, summary)
}


func UpdateSummary(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	dataClient := dataServiceContext.DataClient()

	//details := request.DetailsFromContext(req.Context())
	/*if details == nil {
		request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
		return
	}*/

	responder := request.MustNewResponder(res, req)

	id := req.PathParam("userId")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	summary, err := dataClient.UpdateSummary(req.Context(), id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, summary)
}


func GetAgedSummaries(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	dataClient := dataServiceContext.DataClient()

	//details := request.DetailsFromContext(req.Context())
	/*if details == nil {
		request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
		return
	}*/

	responder := request.MustNewResponder(res, req)

	minutesStr := req.PathParam("minutes")
    if minutesStr == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("minutes"))
		return
	}

    minutes64, err := strconv.ParseUint(minutesStr, 10, 32)
    if err != nil {
        responder.Error(http.StatusBadRequest, request.ErrorParameterInvalid("minutes"))
		return
    }
    minutes := uint(minutes64)

	summaries, err := dataClient.GetAgedSummaries(req.Context(), minutes)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, summaries)
}
