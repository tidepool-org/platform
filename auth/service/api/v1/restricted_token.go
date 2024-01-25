package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
)

func (r *Router) RestrictedTokensRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/users/:userId/restricted_tokens", api.RequireServer(r.ListUserRestrictedTokens)),
		rest.Post("/v1/users/:userId/restricted_tokens", api.RequireAuth(r.CreateUserRestrictedToken)),
		rest.Delete("/v1/users/:userId/restricted_tokens", api.RequireServer(r.DeleteAllRestrictedTokens)),
		rest.Get("/v1/restricted_tokens/:id", api.RequireServer(r.GetRestrictedToken)),
		rest.Put("/v1/restricted_tokens/:id", api.RequireServer(r.UpdateRestrictedToken)),
		rest.Delete("/v1/restricted_tokens/:id", api.RequireAuth(r.DeleteRestrictedToken)),
	}
}

func (r *Router) ListUserRestrictedTokens(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	filter := auth.NewRestrictedTokenFilter()
	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	restrictedTokens, err := r.AuthClient().ListUserRestrictedTokens(req.Context(), userID, filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, restrictedTokens)
}

func (r *Router) CreateUserRestrictedToken(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.GetAuthDetails(req.Context())

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	if !details.IsService() && details.UserID() != userID {
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	create := auth.NewRestrictedTokenCreate()
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	restrictedToken, err := r.AuthClient().CreateUserRestrictedToken(req.Context(), userID, create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusCreated, restrictedToken)
}

func (r *Router) DeleteAllRestrictedTokens(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	if err := r.AuthClient().DeleteAllRestrictedTokens(req.Context(), userID); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusNoContent)
}

func (r *Router) GetRestrictedToken(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	restrictedToken, err := r.AuthClient().GetRestrictedToken(req.Context(), id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if restrictedToken == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	responder.Data(http.StatusOK, restrictedToken)
}

func (r *Router) UpdateRestrictedToken(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	update := auth.NewRestrictedTokenUpdate()
	if err := request.DecodeRequestBody(req.Request, update); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	restrictedToken, err := r.AuthClient().UpdateRestrictedToken(req.Context(), id, update)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if restrictedToken == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	responder.Data(http.StatusOK, restrictedToken)
}

func (r *Router) DeleteRestrictedToken(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.GetAuthDetails(req.Context())

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	restrictedToken, err := r.AuthClient().GetRestrictedToken(req.Context(), id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if restrictedToken == nil {
		responder.Empty(http.StatusOK)
		return
	}

	if !details.IsService() && details.UserID() != restrictedToken.UserID {
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	err = r.AuthClient().DeleteRestrictedToken(req.Context(), id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusOK)
}
