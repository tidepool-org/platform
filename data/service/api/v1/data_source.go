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
		dataService.MakeRoute("DELETE", "/v1/users/:userId/data_sources", Authenticate(DeleteAllSources)),
		dataService.MakeRoute("GET", "/v1/data_sources/:id", Authenticate(GetSource)),
		dataService.MakeRoute("PUT", "/v1/data_sources/:id", Authenticate(UpdateSource)),
		dataService.MakeRoute("DELETE", "/v1/data_sources/:id", Authenticate(DeleteSource)),
	}
}

// func (r *Router) ListSources(res rest.ResponseWriter, req *rest.Request) {

// ListSources godoc
// @Summary List data sources
// @ID platform-data-api-ListSources
// @Produce json
// @Param userId path string true "user ID"
// @Param page query int false "When using pagination, page number" default(0)
// @Param size query int false "When using pagination, number of elements by page, 1<size<1000" minimum(1) maximum(1000) default(100)
// @Param X-Tidepool-Service-Secret header string false "The platform-data service secret"
// @Param X-Tidepool-Session-Token header string false "A tidepool session token"
// @Param restricted_token header string false "A tidepool restricted token"
// @Param Authorization header string false "A tidepool authorization token"
// @Success 200 {array} source.Source "Array of data sources"
// @Failure 400 {object} service.Error "Bad request (userId is missing)"
// @Failure 401 {object} service.Error "Not authenticated"
// @Failure 403 {object} service.Error "Forbidden (only services and owner can access data sources)"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/users/:userId/data_sources [get]
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

// CreateSource godoc
// @Summary Create a data source
// @ID platform-data-api-CreateSource
// @Accept json
// @Produce json
// @Param userId path string true "user ID"
// @Param source.Create body source.Create true "The source to create"
// @Param X-Tidepool-Service-Secret header string false "The platform-data service secret"
// @Param X-Tidepool-Session-Token header string false "A tidepool session token"
// @Param restricted_token header string false "A tidepool restricted token"
// @Param Authorization header string false "A tidepool authorization token"
// @Success 201 {object} source.Source "The created source"
// @Failure 400 {object} service.Error "Bad request (userId is missing)"
// @Failure 401 {object} service.Error "Not authenticated"
// @Failure 403 {object} service.Error "Forbidden (only services can create a source)"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/users/:userId/data_sources [post]
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
// func (r *Router) DeleteAllSources(res rest.ResponseWriter, req *rest.Request) {

// DeleteAllSources godoc
// @Summary Create a data source
// @ID platform-data-api-DeleteAllSources
// @Produce json
// @Param userId path string true "user ID"
// @Param X-Tidepool-Service-Secret header string false "The platform-data service secret"
// @Param X-Tidepool-Session-Token header string false "A tidepool session token"
// @Param restricted_token header string false "A tidepool restricted token"
// @Param Authorization header string false "A tidepool authorization token"
// @Success 204 "Empty content"
// @Failure 400 {object} service.Error "Bad request (userId is missing)"
// @Failure 401 {object} service.Error "Not authenticated"
// @Failure 403 {object} service.Error "Forbidden (only services can delete a source)"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/users/:userId/data_sources [delete]
func DeleteAllSources(dataServiceContext dataService.Context) {
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

	err := dataServiceContext.DataSourceClient().DeleteAll(req.Context(), userID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusNoContent)
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) GetSource(res rest.ResponseWriter, req *rest.Request) {

// GetSource godoc
// @Summary Get a data source
// @ID platform-data-api-GetSource
// @Produce json
// @Param id path string true "The data source ID"
// @Param X-Tidepool-Service-Secret header string false "The platform-data service secret"
// @Param X-Tidepool-Session-Token header string false "A tidepool session token"
// @Param restricted_token header string false "A tidepool restricted token"
// @Param Authorization header string false "A tidepool authorization token"
// @Success 200 {object} source.Source "A data source"
// @Failure 400 {object} service.Error "Bad request (id is missing)"
// @Failure 401 {object} service.Error "Not authenticated"
// @Failure 403 {object} service.Error "Forbidden (only services and owner can access data sources)"
// @Failure 404 {object} service.Error "Not found"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/data_sources/:id [get]
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

// UpdateSource godoc
// @Summary Update a data source
// @ID platform-data-api-UpdateSource
// @Produce json
// @Accept json
// @Param id path string true "The data source ID"
// @Param revision query integer false "Only perform the update if the current data source revision is the one specified"
// @Param X-Tidepool-Service-Secret header string false "The platform-data service secret"
// @Param X-Tidepool-Session-Token header string false "A tidepool session token"
// @Param restricted_token header string false "A tidepool restricted token"
// @Param Authorization header string false "A tidepool authorization token"
// @Param DataSourceUpdate body source.Update true "The update fields of the data source"
// @Success 200 {object} source.Source "The data source updated"
// @Failure 400 {object} service.Error "Bad request (id is missing, bad revision value)"
// @Failure 401 {object} service.Error "Not authenticated"
// @Failure 403 {object} service.Error "Forbidden (only services can update a data source)"
// @Failure 404 {object} service.Error "Not found"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/data_sources/:id [put]
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

// DeleteSource godoc
// @Summary Delete a data source
// @ID platform-data-api-DeleteSource
// @Param id path string true "The data source ID"
// @Param revision query integer false "Only perform the update if the current data source revision is the one specified"
// @Param X-Tidepool-Service-Secret header string false "The platform-data service secret"
// @Param X-Tidepool-Session-Token header string false "A tidepool session token"
// @Param restricted_token header string false "A tidepool restricted token"
// @Param Authorization header string false "A tidepool authorization token"
// @Success 200 "Empty content"
// @Failure 400 {object} service.Error "Bad request (id is missing, bad revision value)"
// @Failure 401 {object} service.Error "Not authenticated"
// @Failure 403 {object} service.Error "Forbidden (only services can update a data source)"
// @Failure 404 {object} service.Error "Not found"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/data_sources/:id [delete]
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
