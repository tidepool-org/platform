package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	serviceApi "github.com/tidepool-org/platform/service/api"
	twiistProvider "github.com/tidepool-org/platform/twiist/provider"
)

func (r *Router) ProviderSessionsRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/users/:userId/provider_sessions", serviceApi.RequireServer(r.ListUserProviderSessions)),
		rest.Post("/v1/users/:userId/provider_sessions", serviceApi.RequireServer(r.CreateUserProviderSession)),
		rest.Delete("/v1/users/:userId/provider_sessions", serviceApi.RequireServer(r.DeleteUserProviderSessions)),
		rest.Get("/v1/provider_sessions/:id", serviceApi.RequireServer(r.GetProviderSession)),
		rest.Put("/v1/provider_sessions/:id", serviceApi.RequireServer(r.UpdateProviderSession)),
		rest.Delete("/v1/provider_sessions/:id", serviceApi.RequireServer(r.DeleteProviderSession)),

		// TODO: Temporary endpoint for deleting provider sessions given a twiist tidepool link id
		rest.Delete("/v1/partners/twiist/links/:tidepoolLinkId", serviceApi.RequireAuth(r.DeleteProviderSessionByTidepoolLinkID)),
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

func (r *Router) DeleteUserProviderSessions(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	if err := r.AuthClient().DeleteUserProviderSessions(req.Context(), userID); err != nil {
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
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	filter := auth.ProviderSessionFilter{
		Type:       pointer.FromString(oauthProvider.ProviderType),
		Name:       pointer.FromString(twiistProvider.ProviderName),
		ExternalID: pointer.FromString(tidepoolLinkID),
	}
	if err := r.AuthServiceClient().DeleteAllProviderSessions(req.Context(), filter); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusOK)
}
