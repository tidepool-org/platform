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
		rest.Get("/v1/users/:userId/profile", api.RequireUser(r.GetProfile)),
		rest.Get("/v1/users/:userId/legacy_profile", api.RequireUser(r.GetLegacyProfile)),
		// The following modification routes required custodian access in seagull, but I'm not sure that's quite right - it seems it should be if the user can modify the userId.
		rest.Put("/v1/users/:userId/profile", r.requireWriteAccess("userId", r.UpdateProfile)),
		rest.Put("/v1/users/:userId/legacy_profile", r.requireWriteAccess("userId", r.UpdateLegacyProfile)),
		rest.Delete("/v1/users/:userId/profile", r.requireWriteAccess("userId", r.DeleteProfile)),
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
	if user == nil || user.Profile == nil {
		responder.Empty(http.StatusNotFound)
		return
	}

	responder.Data(http.StatusOK, user.Profile)
}

func (r *Router) GetLegacyProfile(res rest.ResponseWriter, req *rest.Request) {
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
	if user == nil || user.Profile == nil {
		responder.Empty(http.StatusNotFound)
		return
	}

	responder.Data(http.StatusOK, user.Profile.ToLegacyProfile())
}

func (r *Router) UpdateProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	profile := &user.UserProfile{}
	if err := request.DecodeRequestBody(req.Request, profile); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	r.updateProfile(res, req, profile)
}

func (r *Router) UpdateLegacyProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	profile := &user.LegacyUserProfile{}
	if err := request.DecodeRequestBody(req.Request, profile); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	r.updateProfile(res, req, profile.ToUserProfile())
}

func (r *Router) updateProfile(res rest.ResponseWriter, req *rest.Request, profile *user.UserProfile) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	userID := req.PathParam("userId")
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
