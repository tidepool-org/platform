package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	serviceApi "github.com/tidepool-org/platform/service/api"
	"github.com/tidepool-org/platform/twiist"
)

func (r *Router) ProviderSessionsRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/users/:userId/provider_sessions", serviceApi.RequireServer(r.CreateUserProviderSession)),    // DEPRECATED: Use CreateProviderSession
		rest.Delete("/v1/users/:userId/provider_sessions", serviceApi.RequireServer(r.DeleteUserProviderSessions)), // DEPRECATED: Use DeleteProviderSessions

		rest.Get("/v1/provider_sessions", serviceApi.RequireServer(r.ListProviderSessions)),
		rest.Delete("/v1/provider_sessions", serviceApi.RequireServer(r.DeleteProviderSessions)),

		rest.Post("/v1/provider_sessions", serviceApi.RequireServer(r.CreateProviderSession)),
		rest.Get("/v1/provider_sessions/:id", serviceApi.RequireServer(r.GetProviderSession)),
		rest.Put("/v1/provider_sessions/:id", serviceApi.RequireServer(r.UpdateProviderSession)),
		rest.Delete("/v1/provider_sessions/:id", serviceApi.RequireServer(r.DeleteProviderSession)),

		// TODO: Temporary endpoint for deleting provider sessions given a twiist tidepool link id
		rest.Delete("/v1/partners/twiist/links/:tidepoolLinkId", serviceApi.RequireAuth(r.DeleteProviderSessionByTidepoolLinkID)),
	}
}

// DEPRECATED: Use CreateProviderSession
func (r *Router) CreateUserProviderSession(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	create := auth.NewProviderSessionCreate()
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	create.UserID = userID

	providerSession, err := r.AuthClient().CreateProviderSession(req.Context(), create)
	if err != nil {
		responder.InternalServerError(err)
		return
	}

	responder.Data(http.StatusCreated, providerSession)
}

// DEPRECATED: Use DeleteProviderSessions
func (r *Router) DeleteUserProviderSessions(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	filter := &auth.ProviderSessionFilter{UserID: &userID}
	if err := r.AuthClient().DeleteProviderSessions(req.Context(), filter); err != nil {
		responder.InternalServerError(err)
		return
	}

	responder.Empty(http.StatusNoContent)
}

func (r *Router) ListProviderSessions(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	filter := auth.NewProviderSessionFilter()
	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	providerSessions, err := r.AuthClient().ListProviderSessions(req.Context(), filter, pagination)
	if err != nil {
		responder.InternalServerError(err)
		return
	}

	responder.Data(http.StatusOK, providerSessions)
}

func (r *Router) DeleteProviderSessions(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	filter := auth.NewProviderSessionFilter()
	if err := request.DecodeRequestQuery(req.Request, filter); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	if err := r.AuthClient().DeleteProviderSessions(req.Context(), filter); err != nil {
		responder.InternalServerError(err)
		return
	}

	responder.Empty(http.StatusNoContent)
}

func (r *Router) CreateProviderSession(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	create := auth.NewProviderSessionCreate()
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	providerSession, err := r.AuthClient().CreateProviderSession(req.Context(), create)
	if err != nil {
		responder.InternalServerError(err)
		return
	}

	responder.Data(http.StatusCreated, providerSession)
}

func (r *Router) GetProviderSession(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	providerSession, err := r.AuthClient().GetProviderSession(req.Context(), id)
	if err != nil {
		responder.InternalServerError(err)
		return
	} else if providerSession == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	responder.Data(http.StatusOK, providerSession)
}

func (r *Router) UpdateProviderSession(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	update := auth.NewProviderSessionUpdate()
	if err := request.DecodeRequestBody(req.Request, update); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	providerSession, err := r.AuthClient().UpdateProviderSession(req.Context(), id, update)
	if err != nil {
		responder.InternalServerError(err)
		return
	} else if providerSession == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	responder.Data(http.StatusOK, providerSession)
}

func (r *Router) DeleteProviderSession(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	err := r.AuthClient().DeleteProviderSession(req.Context(), id)
	if err != nil {
		responder.InternalServerError(err)
		return
	}

	responder.Empty(http.StatusOK)
}

// TODO: Temporary endpoint for deleting provider sessions given a twiist tidepool link id
func (r *Router) DeleteProviderSessionByTidepoolLinkID(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	tidepoolLinkID := req.PathParams["tidepoolLinkId"]
	if tidepoolLinkID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("tidepoolLinkId"))
		return
	}

	// Authorize the service account
	authDetails := request.GetAuthDetails(req.Context())
	if !authDetails.IsService() && !r.TwiistServiceAccountAuthorizer().IsServiceAccountAuthorized(authDetails.UserID()) {
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	filter := &auth.ProviderSessionFilter{
		Type:       pointer.FromString(oauth.ProviderType),
		Name:       pointer.FromString(twiist.ProviderName),
		ExternalID: pointer.FromString(tidepoolLinkID),
	}
	if err := r.AuthClient().DeleteProviderSessions(req.Context(), filter); err != nil {
		responder.InternalServerError(err)
		return
	}

	responder.Empty(http.StatusOK)
}
