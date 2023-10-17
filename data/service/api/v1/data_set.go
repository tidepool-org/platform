package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/data"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
)

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) DataSetsRoutes() []*rest.Route {
// 	return []*rest.Route{
// 		rest.Get("/v1/users/:userId/data_sets", api.Require(r.ListUserDataSets)),
// 		rest.Get("/v1/data_sets/:id", api.Require(r.GetDataSet)),
// 	}
// }

func DataSetsRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.MakeRoute("GET", "/v1/users/:userId/data_sets", EnforceAuthentication(ListUserDataSets)),
		dataService.MakeRoute("GET", "/v1/data_sets/:dataSetId", EnforceAuthentication(GetDataSet)),
	}
}

// func (r *Router) ListUserDataSets(res rest.ResponseWriter, req *rest.Request) {

func ListUserDataSets(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	dataClient := dataServiceContext.DataClient()

	details := request.GetAuthDetails(req.Context())
	if details == nil {
		request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
		return
	}
	// TODO: END: Update to new service paradigm

	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	// FUTURE: Refactor for global usage
	if !details.IsService() && details.UserID() != userID {
		permissions, err := dataServiceContext.PermissionClient().GetUserPermissions(req.Context(), details.UserID(), userID)
		if err != nil {
			if request.IsErrorUnauthorized(err) {
				responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
			} else {
				responder.Error(http.StatusInternalServerError, err)
			}
			return
		}
		_, custodianPermission := permissions[permission.Custodian]
		_, uploadPermission := permissions[permission.Write]
		_, viewPermission := permissions[permission.Read]
		if !custodianPermission && !uploadPermission && !viewPermission {
			responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
			return
		}
	}

	filter := data.NewDataSetFilter()
	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	dataSets, err := dataClient.ListUserDataSets(req.Context(), userID, filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, dataSets)
}

// func (r *Router) GetDataSet(res rest.ResponseWriter, req *rest.Request) {

func GetDataSet(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	dataClient := dataServiceContext.DataClient()

	details := request.GetAuthDetails(req.Context())
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

	if !details.IsService() && details.UserID() != *dataSet.UserID {
		request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	responder.Data(http.StatusOK, dataSet)
}
