package v1

import (
	stdErrs "errors"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/request"
	structValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

func (r *Router) ProfileRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/users/:userId/profile", r.requireMembership("userId", r.GetProfile)),
		rest.Get("/v1/users/legacy/:userId/profile", r.requireMembership("userId", r.GetLegacyProfile)),
		rest.Put("/v1/users/:userId/profile", r.requireCustodian("userId", r.UpdateProfile)),
		rest.Put("/v1/users/legacy/:userId/profile", r.requireCustodian("userId", r.UpdateLegacyProfile)),
		rest.Post("/v1/users/:userId/profile", r.requireCustodian("userId", r.UpdateProfile)),
		rest.Post("/v1/users/legacy/:userId/profile", r.requireCustodian("userId", r.UpdateLegacyProfile)),
		rest.Delete("/v1/users/:userId/profile", r.requireCustodian("userId", r.DeleteProfile)),
		rest.Delete("/v1/users/legacy/:userId/profile", r.requireCustodian("userId", r.DeleteProfile)),
	}
}

func (r *Router) GetProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	userID := req.PathParam("userId")

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
	userID := req.PathParam("userId")

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
