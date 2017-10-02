package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/data"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
)

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) DataSourcesRoutes() []*rest.Route {
// 	return []*rest.Route{
// 		rest.Get("/v1/users/:userId/data_sources", api.Require(r.ListUserDataSources)),
// 		rest.Post("/v1/users/:userId/data_sources", api.RequireServer(r.CreateUserDataSource)),
// 		rest.Get("/v1/data_sources/:id", api.Require(r.GetDataSource)),
// 		rest.Put("/v1/data_sources/:id", api.RequireServer(r.UpdateDataSource)),
// 		rest.Delete("/v1/data_sources/:id", api.RequireServer(r.DeleteDataSource)),
// 	}
// }

func DataSourcesRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.MakeRoute("GET", "/v1/users/:userId/data_sources", Authenticate(ListUserDataSources)),
		dataService.MakeRoute("POST", "/v1/users/:userId/data_sources", Authenticate(CreateUserDataSource)),
		dataService.MakeRoute("GET", "/v1/data_sources/:id", Authenticate(GetDataSource)),
		dataService.MakeRoute("PUT", "/v1/data_sources/:id", Authenticate(UpdateDataSource)),
		dataService.MakeRoute("DELETE", "/v1/data_sources/:id", Authenticate(DeleteDataSource)),
	}
}

// func (r *Router) ListUserDataSources(res rest.ResponseWriter, req *rest.Request) {

func ListUserDataSources(dataServiceContext dataService.Context) {
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

	if !details.IsService() && details.UserID() != userID {
		request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	filter := data.NewDataSourceFilter()
	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	dataSources, err := dataClient.ListUserDataSources(req.Context(), userID, filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, dataSources)
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) CreateUserDataSource(res rest.ResponseWriter, req *rest.Request) {

func CreateUserDataSource(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	dataClient := dataServiceContext.DataClient()

	details := request.DetailsFromContext(req.Context())
	if details == nil {
		request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
		return
	} else if !details.IsService() {
		request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}
	// TODO: END: Update to new service paradigm

	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	create := data.NewDataSourceCreate()
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	dataSource, err := dataClient.CreateUserDataSource(req.Context(), userID, create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusCreated, dataSource)
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) GetDataSource(res rest.ResponseWriter, req *rest.Request) {

func GetDataSource(dataServiceContext dataService.Context) {
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

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	dataSource, err := dataClient.GetDataSource(req.Context(), id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if dataSource == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	if !details.IsService() && details.UserID() != dataSource.UserID {
		request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	responder.Data(http.StatusOK, dataSource)
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) UpdateDataSource(res rest.ResponseWriter, req *rest.Request) {

func UpdateDataSource(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	dataClient := dataServiceContext.DataClient()

	details := request.DetailsFromContext(req.Context())
	if details == nil {
		request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
		return
	} else if !details.IsService() {
		request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}
	// TODO: END: Update to new service paradigm

	responder := request.MustNewResponder(res, req)

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	update := data.NewDataSourceUpdate()
	if err := request.DecodeRequestBody(req.Request, update); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	dataSource, err := dataClient.UpdateDataSource(req.Context(), id, update)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if dataSource == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	responder.Data(http.StatusOK, dataSource)
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) DeleteDataSource(res rest.ResponseWriter, req *rest.Request) {

func DeleteDataSource(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	dataClient := dataServiceContext.DataClient()

	details := request.DetailsFromContext(req.Context())
	if details == nil {
		request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
		return
	} else if !details.IsService() {
		request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}
	// TODO: END: Update to new service paradigm

	responder := request.MustNewResponder(res, req)

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	err := dataClient.DeleteDataSource(req.Context(), id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusOK)
}
