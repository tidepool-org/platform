package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/request"
)

func GetHasAnyData(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	responder := request.MustNewResponder(res, req)
	dataClient := dataServiceContext.DataClient()
	userID := req.PathParam("userId")

	has, err := dataClient.HasAnyData(req.Context(), userID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}
	if has {
		responder.Empty(http.StatusCreated)
		return
	}

	dataSourceClient := dataServiceContext.DataSourceClient()
	has, err = dataSourceClient.HasAnyData(req.Context(), userID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}
	if has {
		responder.Empty(http.StatusCreated)
		return
	}
	responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(userID))
}
