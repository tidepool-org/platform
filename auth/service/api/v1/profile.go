package v1

import (
	stdErrs "errors"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
	structValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

func (r *Router) ProfileRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/profiles/:userId", api.RequireUser(r.GetProfile)),
		rest.Put("/v1/profiles/:userId", r.requireUserHasCustodian("userId", r.UpdateProfile)),
		rest.Delete("/v1/profiles/:userId", r.requireUserHasCustodian("userId", r.DeleteProfile)),
	}
}

func (r *Router) GetProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	details := request.GetAuthDetails(ctx)
	userID := req.PathParam("userId")
	hasPerms, err := r.PermissionsClient().HasMembershipRelationship(ctx, details.UserID(), userID)
	if err != nil {
		responder.InternalServerError(err)
		return
	}
	if !hasPerms {
		responder.Empty(http.StatusForbidden)
		return
	}

	user, err := r.UserAccessor().FindUserById(ctx, userID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	responder.Data(http.StatusOK, user.Profile)
}

func (r *Router) UpdateProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	userID := req.PathParam("userId")

	profile := &user.UserProfile{}
	if err := request.DecodeRequestBody(req.Request, profile); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	if err := structValidator.New().Validate(profile); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	err := r.UserAccessor().UpdateUserProfile(ctx, userID, profile)
	if stdErrs.Is(err, user.ErrUserNotFound) {
		responder.Empty(http.StatusNotFound)
		return
	}
	if err != nil {
		responder.InternalServerError(err)
		return
	}
	responder.Empty(http.StatusOK)
}

func (r *Router) DeleteProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	userID := req.PathParam("userId")

	err := r.UserAccessor().DeleteUserProfile(ctx, userID)
	if stdErrs.Is(err, user.ErrUserNotFound) {
		responder.Empty(http.StatusNotFound)
		return
	}
	if err != nil {
		responder.InternalServerError(err)
		return
	}
	responder.Empty(http.StatusOK)
}
