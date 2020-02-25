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
		dataService.MakeRoute("GET", "/v1/users/:userId/data_sets", Authenticate(ListUserDataSets)),
		dataService.MakeRoute("GET", "/v1/data_sets/:dataSetId", Authenticate(GetDataSet)),
	}
}

// func (r *Router) ListUserDataSets(res rest.ResponseWriter, req *rest.Request) {

// ListUserDataSets godoc
// @Summary List the user's datasets
// @Produce json
// @Param userId path string true "user ID"
// @Param page query int false "When using pagination, page number" default(0)
// @Param size query int false "When using pagination, number of elements by page, 1<size<1000" minimum(1) maximum(1000) default(100)
// @Param X-Tidepool-Service-Secret header string false "The platform-data service secret"
// @Param X-Tidepool-Session-Token header string false "A tidepool session token"
// @Param restricted_token header string false "A tidepool restricted token"
// @Param Authorization header string false "A tidepool authorization token"
// @Success 200 {array} data.DataSet "Array of data sets"
// @Failure 400 {object} service.Error "Bad request (userId is missing)"
// @Failure 401 {object} service.Error "Not authenticated"
// @Failure 403 {object} service.Error "Forbidden"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/users/:userId/data_sets [get]
func ListUserDataSets(dataServiceContext dataService.Context) {
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

// GetDataSet godoc
// @Summary Get one dataset
// @Produce json
// @Param dataSetId path string true "dataSet ID"
// @Param X-Tidepool-Service-Secret header string false "The platform-data service secret"
// @Param X-Tidepool-Session-Token header string false "A tidepool session token"
// @Param restricted_token header string false "A tidepool restricted token"
// @Param Authorization header string false "A tidepool authorization token"
// @Success 200 {object} data.DataSet "The requested data set"
// @Failure 400 {object} service.Error "Bad request (userId is missing)"
// @Failure 401 {object} service.Error "Not authenticated"
// @Failure 403 {object} service.Error "Forbidden"
// @Failure 404 {object} service.Error "Dataset not found"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/data_sets/:dataSetId [get]
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

	if !details.IsService() && details.UserID() != *dataSet.UserID {
		request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	responder.Data(http.StatusOK, dataSet)
}
