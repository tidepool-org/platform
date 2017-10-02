package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/request"
)

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) DataSetsRoutes() []*rest.Route {
// 	return []*rest.Route{
// 		rest.Get("/v1/data_sets/:id", api.Require(r.GetDataSet)),
// 	}
// }

func DataSetsRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.MakeRoute("GET", "/v1/data_sets/:dataSetId", Authenticate(GetDataSet)),
	}
}

// func (r *Router) GetDataSet(res rest.ResponseWriter, req *rest.Request) {

func GetDataSet(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	dataClient := dataServiceContext.DataClient()

	details := request.DetailsFromContext(req.Context())
	if details == nil {
		request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
		return
	}
	// TODO: END: Update to new service paradigm

	responder := request.MustNewResponder(res, req)

	id := req.PathParam("dataSetId") // TODO: Use "id"
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	dataSet, err := dataClient.GetDataSet(req.Context(), id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if dataSet == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	if !details.IsService() && details.UserID() != dataSet.UserID {
		request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	responder.Data(http.StatusOK, dataSet)
}
