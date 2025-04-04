package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/errors"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/twiist/provider"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
)

func (r *Router) ProviderSessionsRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/users/:userId/provider_sessions", api.RequireServer(r.ListUserProviderSessions)),
		rest.Post("/v1/users/:userId/provider_sessions", api.RequireServer(r.CreateUserProviderSession)),
		rest.Delete("/v1/users/:userId/provider_sessions", api.RequireServer(r.DeleteAllProviderSessions)),
		rest.Get("/v1/provider_sessions/:id", api.RequireServer(r.GetProviderSession)),
		rest.Put("/v1/provider_sessions/:id", api.RequireServer(r.UpdateProviderSession)),
		rest.Delete("/v1/provider_sessions/:id", api.RequireServer(r.DeleteProviderSession)),

		// Temporary endpoint for deleting provider sessions given a twiist tidepool link id
		rest.Delete("/v1/partners/twiist/links/:tidepoolLinkId", api.RequireAuth(r.DeleteProviderSessionByTidepoolLinkID)),
	}
}

func (r *Router) ListUserProviderSessions(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	filter := auth.NewProviderSessionFilter()
	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	providerSessions, err := r.AuthClient().ListUserProviderSessions(req.Context(), userID, filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, providerSessions)
}

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

	providerSession, err := r.AuthClient().CreateUserProviderSession(req.Context(), userID, create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusCreated, providerSession)
}

func (r *Router) DeleteAllProviderSessions(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	if err := r.AuthClient().DeleteAllProviderSessions(req.Context(), userID); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusNoContent)
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
		responder.Error(http.StatusInternalServerError, err)
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
		responder.Error(http.StatusInternalServerError, err)
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
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusOK)
}

func (r *Router) DeleteProviderSessionByTidepoolLinkID(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	tidepoolLinkID := req.PathParams["tidepoolLinkId"]
	if tidepoolLinkID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("tidepoolLinkId"))
		return
	}

	// Authorize the service account
	authDetails := request.GetAuthDetails(req.Context())
	if !authDetails.IsService() && !r.TwiistServiceAccountAuthorizer().IsAuthorized(authDetails.UserID()) {
		responder.Error(http.StatusForbidden, errors.New("auth token is not authorized for requested action"))
		return
	}

	filter := auth.ProviderSessionFilter{
		Type:       pointer.FromString(oauthProvider.ProviderType),
		Name:       pointer.FromString(provider.ProviderName),
		ExternalID: pointer.FromString(tidepoolLinkID),
	}
	err := r.AuthClient().DeleteAllProviderSessionsByExternalID(req.Context(), filter)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusOK)
}
