package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
)

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) SourcesRoutes() []*rest.Route {
// 	return []*rest.Route{
// 		rest.Get("/v1/users/:userId/data_sources", api.Require(r.ListSources)),
// 		rest.Post("/v1/users/:userId/data_sources", api.RequireServer(r.CreateSource)),
// 		rest.Get("/v1/data_sources/:id", api.Require(r.GetSource)),
// 		rest.Put("/v1/data_sources/:id", api.RequireServer(r.UpdateSource)),
// 		rest.Delete("/v1/data_sources/:id", api.RequireServer(r.DeleteSource)),
// 	}
// }

func SourcesRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.MakeRoute("GET", "/v1/users/:userId/data_sources", Authenticate(ListSources)),
		dataService.MakeRoute("POST", "/v1/users/:userId/data_sources", Authenticate(CreateSource)),
		dataService.MakeRoute("GET", "/v1/data_sources/:id", Authenticate(GetSource)),
		dataService.MakeRoute("PUT", "/v1/data_sources/:id", Authenticate(UpdateSource)),
		dataService.MakeRoute("DELETE", "/v1/data_sources/:id", Authenticate(DeleteSource)),
	}
}

// func (r *Router) ListSources(res rest.ResponseWriter, req *rest.Request) {

func ListSources(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

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

	filter := dataSource.NewFilter()
	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	sources, err := dataServiceContext.DataSourceClient().List(req.Context(), userID, filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, sources)
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) CreateSource(res rest.ResponseWriter, req *rest.Request) {

func CreateSource(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

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

	create := dataSource.NewCreate()
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	source, err := dataServiceContext.DataSourceClient().Create(req.Context(), userID, create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusCreated, source)
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) GetSource(res rest.ResponseWriter, req *rest.Request) {

func GetSource(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

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

	source, err := dataServiceContext.DataSourceClient().Get(req.Context(), id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if source == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	if !details.IsService() && details.UserID() != *source.UserID {
		request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	responder.Data(http.StatusOK, source)
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) UpdateSource(res rest.ResponseWriter, req *rest.Request) {

func UpdateSource(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

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

	condition := request.NewCondition()
	if err := request.DecodeRequestQuery(req.Request, condition); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	update := dataSource.NewUpdate()
	if err := request.DecodeRequestBody(req.Request, update); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	source, err := dataServiceContext.DataSourceClient().Update(req.Context(), id, condition, update)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if source == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, condition.Revision))
		return
	}

	responder.Data(http.StatusOK, source)
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) DeleteSource(res rest.ResponseWriter, req *rest.Request) {

func DeleteSource(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

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

	condition := request.NewCondition()
	if err := request.DecodeRequestQuery(req.Request, condition); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	deleted, err := dataServiceContext.DataSourceClient().Delete(req.Context(), id, condition)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if !deleted {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, condition.Revision))
		return
	}

	responder.Empty(http.StatusOK)
}
