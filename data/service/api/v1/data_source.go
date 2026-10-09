package v1

import (
	"net/http"
	"slices"

	"github.com/tidepool-org/platform/data"
	dataService "github.com/tidepool-org/platform/data/service"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	serviceApi "github.com/tidepool-org/platform/service/api"
	"github.com/tidepool-org/platform/structure"
)

func SourcesRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.Get("/v1/users/:userId/data_sources", ListSources, serviceApi.RequireAuth),
		dataService.Post("/v1/users/:userId/data_sources", CreateSource, serviceApi.RequireAuth),
		dataService.Delete("/v1/users/:userId/data_sources", DeleteAllSources, serviceApi.RequireAuth),
		dataService.Get("/v1/data_sources/:id", GetSource, serviceApi.RequireAuth),
		dataService.Put("/v1/data_sources/:id", UpdateSource, serviceApi.RequireAuth),
		dataService.Delete("/v1/data_sources/:id", DeleteSource, serviceApi.RequireAuth),

		dataService.Get("/v1/provider_sessions/:providerSessionId/data_source", GetSourceFromProviderSession, serviceApi.RequireAuth),
	}
}

// func (r *Router) ListSources(res rest.ResponseWriter, req *rest.Request) {

func ListSources(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

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
		responder.InternalServerError(err)
		return
	}

	responder.Data(http.StatusOK, newSourceArrayDEPRECATED(sources))
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) CreateSource(res rest.ResponseWriter, req *rest.Request) {

func CreateSource(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	details := request.GetAuthDetails(req.Context())
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
		responder.InternalServerError(err)
		return
	}

	responder.Data(http.StatusCreated, newSourceDEPRECATED(source))
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) DeleteAllSources(res rest.ResponseWriter, req *rest.Request) {

func DeleteAllSources(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	details := request.GetAuthDetails(req.Context())
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
		responder.InternalServerError(err)
		return
	}

	responder.Empty(http.StatusNoContent)
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) GetSource(res rest.ResponseWriter, req *rest.Request) {

func GetSource(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	details := request.GetAuthDetails(req.Context())
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
		responder.InternalServerError(err)
		return
	} else if source == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	if !details.IsService() && details.UserID() != source.UserID {
		request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	responder.Data(http.StatusOK, newSourceDEPRECATED(source))
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) UpdateSource(res rest.ResponseWriter, req *rest.Request) {

func UpdateSource(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	details := request.GetAuthDetails(req.Context())
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

	update := &UpdateDEPRECATED{}
	if err := request.DecodeRequestBody(req.Request, update); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	if update.DataSetID != nil {
		source, err := dataServiceContext.DataSourceClient().Get(req.Context(), id)
		if err != nil {
			responder.InternalServerError(err)
			return
		} else if source == nil {
			responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
			return
		}
		update.Modernize(source)
	}

	source, err := dataServiceContext.DataSourceClient().Update(req.Context(), id, condition, &update.Update)
	if err != nil {
		responder.InternalServerError(err)
		return
	} else if source == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, condition.Revision))
		return
	}

	responder.Data(http.StatusOK, newSourceDEPRECATED(source))
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) DeleteSource(res rest.ResponseWriter, req *rest.Request) {

func DeleteSource(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	details := request.GetAuthDetails(req.Context())
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
		responder.InternalServerError(err)
		return
	} else if !deleted {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, condition.Revision))
		return
	}

	responder.Empty(http.StatusOK)
}

// TODO: BEGIN: Update to new service paradigm
// func (r *Router) GetSourceFromProviderSession(res rest.ResponseWriter, req *rest.Request) {

func GetSourceFromProviderSession(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	details := request.GetAuthDetails(req.Context())
	if details == nil {
		request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
		return
	} else if !details.IsService() {
		request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}
	// TODO: END: Update to new service paradigm

	responder := request.MustNewResponder(res, req)

	providerSessionID := req.PathParam("providerSessionId")
	if providerSessionID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("providerSessionId"))
		return
	}

	source, err := dataServiceContext.DataSourceClient().GetFromProviderSession(req.Context(), providerSessionID)
	if err != nil {
		responder.InternalServerError(err)
		return
	} else if source == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(providerSessionID))
		return
	}

	responder.Data(http.StatusOK, newSourceDEPRECATED(source))
}

// UpdateDEPRECATED accepts the single dataSetId that services built before dataSetIds returned (2026-10) still
// send, so the services can be deployed in any order. Remove once every service sends dataSetIds.
type UpdateDEPRECATED struct {
	dataSource.Update `bson:",inline"`

	DataSetID *string `json:"dataSetId,omitempty" bson:"dataSetId,omitempty"`
}

func (u *UpdateDEPRECATED) Parse(parser structure.ObjectParser) {
	u.Update.Parse(parser)

	u.DataSetID = parser.String("dataSetId")
}

func (u *UpdateDEPRECATED) Validate(validator structure.Validator) {
	u.Update.Validate(validator)

	if dataSetIDValidator := validator.String("dataSetId", u.DataSetID); u.DataSetIDs != nil {
		dataSetIDValidator.NotExists()
	} else {
		dataSetIDValidator.Using(data.SetIDValidator)
	}
}

func (u *UpdateDEPRECATED) Normalize(normalizer structure.Normalizer) {
	u.Update.Normalize(normalizer)
}

// Modernize turns the single data set id into the data source's data set ids with it as the last, the way the
// sending service would have had them.
func (u *UpdateDEPRECATED) Modernize(source *dataSource.Source) {
	if u.DataSetID != nil && u.DataSetIDs == nil {
		dataSetIDs := pointer.To(pointer.CloneStringArray(source.DataSetIDs))
		if !slices.Contains(dataSetIDs, *u.DataSetID) {
			dataSetIDs = append(dataSetIDs, *u.DataSetID)
		}
		u.DataSetIDs = pointer.From(dataSetIDs)
	}
}

// SourceDEPRECATED adds the single dataSetId that services built before dataSetIds returned (2026-10) still
// read, so the services can be deployed in any order. Remove once every service reads dataSetIds.
type SourceDEPRECATED struct {
	*dataSource.Source

	DataSetID *string `json:"dataSetId,omitempty" bson:"dataSetId,omitempty"`
}

func newSourceDEPRECATED(source *dataSource.Source) *SourceDEPRECATED {
	if source == nil {
		return nil
	}
	return &SourceDEPRECATED{Source: source, DataSetID: source.LastDataSetID()}
}

type SourceArrayDEPRECATED []*SourceDEPRECATED

func newSourceArrayDEPRECATED(sources dataSource.SourceArray) SourceArrayDEPRECATED {
	result := make(SourceArrayDEPRECATED, len(sources))
	for index, source := range sources {
		result[index] = newSourceDEPRECATED(source)
	}
	return result
}

func (s SourceArrayDEPRECATED) Sanitize(details request.AuthDetails) error {
	for _, source := range s {
		if err := source.Sanitize(details); err != nil {
			return err
		}
	}
	return nil
}
