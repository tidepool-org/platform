package v1

import (
	"context"
	stdErrs "errors"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	structValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

func (r *Router) ProfileRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/users/:userId/profile", r.requireMembership("userId", r.GetProfile)),
		rest.Get("/v1/users/:userId/users", r.requireMembership("userId", r.GetUsersWithProfiles)),
		rest.Get("/v1/users/legacy/:userId/profile", r.requireMembership("userId", r.GetLegacyProfile)),
		rest.Put("/v1/users/:userId/profile", r.requireCustodian("userId", r.UpdateProfile)),
		rest.Put("/v1/users/legacy/:userId/profile", r.requireCustodian("userId", r.UpdateLegacyProfile)),
		rest.Post("/v1/users/:userId/profile", r.requireCustodian("userId", r.UpdateProfile)),
		rest.Post("/v1/users/legacy/:userId/profile", r.requireCustodian("userId", r.UpdateLegacyProfile)),
		rest.Delete("/v1/users/:userId/profile", r.requireCustodian("userId", r.DeleteProfile)),
		rest.Delete("/v1/users/legacy/:userId/profile", r.requireCustodian("userId", r.DeleteProfile)),
	}
}

func (r *Router) getProfile(ctx context.Context, userID string) (*user.UserProfile, error) {
	// Until seagull migration is complete use UserProfileAccessor() to get a profile instead of the profile within the user itself.
	profile, err := r.UserProfileAccessor().FindUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, user.ErrUserProfileNotFound
	}
	// Once seagull migration is compelte, we can return
	// the profile attached to the user directly via person.Profile
	// through r.UserAccessor().FindUserProfile
	return profile, nil
}

func (r *Router) GetProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	userID := req.PathParam("userId")
	if r.handledUserNotExists(ctx, responder, userID) {
		return
	}

	profile, err := r.getProfile(ctx, userID)
	if err != nil {
		r.handleProfileErr(responder, err)
		return
	}
	responder.Data(http.StatusOK, profile)
}

func (r *Router) GetUsersWithProfiles(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	targetUserID := req.PathParam("userId")
	if r.handledUserNotExists(ctx, responder, targetUserID) {
		return
	}

	filter := parseUsersQuery(req.URL.Query())
	if !isUsersQueryValid(filter) {
		responder.Error(http.StatusBadRequest, errors.New("unable to parse users query"))
		return
	}
	mergedUserPerms := map[string]*permission.TrustPermissions{}
	trustorPerms, err := r.PermissionsClient().GroupsForUser(ctx, targetUserID)
	if err != nil {
		responder.InternalServerError(err)
		return
	}
	results := make([]*user.User, 0, len(trustorPerms))
	for userID, perms := range trustorPerms {
		if userID == targetUserID {
			// Don't include own user in result
			continue
		}

		mergedUserPerms[userID] = &permission.TrustPermissions{
			TrustorPermissions: &perms,
		}

		if perms.HasReadPermissions() {
			sharedUser, err := r.UserAccessor().FindUserById(ctx, userID)
			if err != nil {
				responder.InternalServerError(err)
				return
			}
			profile, err := r.getProfile(ctx, userID)
			if err != nil && !stdErrs.Is(err, user.ErrUserProfileNotFound) {
				r.handleProfileErr(responder, err)
				return
			}
			sharedUser.Profile = profile
			// Seems no sharedUser.Sanitize call to filter out "protected" fields in seagull except sanitizeUser to remove "passwordExists" field
			perms := perms
			sharedUser.TrustorPermissions = &perms
			if len(perms) == 0 && profile != nil {
				sharedUser.Profile = nil
			}

			results = append(results, sharedUser)
		}
	}
	responder.Data(http.StatusOK, results)
}

func (r *Router) GetLegacyProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	userID := req.PathParam("userId")
	if r.handledUserNotExists(ctx, responder, userID) {
		return
	}

	profile, err := r.getProfile(ctx, userID)
	if err != nil {
		r.handleProfileErr(responder, err)
		return
	}
	responder.Data(http.StatusOK, profile.ToLegacyProfile())
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
	if r.handledUserNotExists(ctx, responder, userID) {
		return
	}
	// Once seagull migration is complete, we can use r.UserAccessor().UpdateUserProfile.
	if err := r.UserProfileAccessor().UpdateUserProfile(ctx, userID, profile); err != nil {
		r.handleProfileErr(responder, err)
		return
	}
	responder.Empty(http.StatusOK)
}

func (r *Router) DeleteProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	userID := req.PathParam("userId")

	err := r.UserProfileAccessor().DeleteUserProfile(ctx, userID)
	if err != nil {
		r.handleProfileErr(responder, err)
		return
	}
	responder.Empty(http.StatusOK)
}

func (r *Router) handleProfileErr(responder *request.Responder, err error) {
	switch {
	case stdErrs.Is(err, user.ErrUserNotFound), stdErrs.Is(err, user.ErrUserProfileNotFound):
		responder.Empty(http.StatusNotFound)
		return
	default:
		responder.InternalServerError(err)
	}
}

func (r *Router) handledUserNotExists(ctx context.Context, responder *request.Responder, userID string) (handled bool) {
	person, err := r.UserAccessor().FindUserById(ctx, userID)
	if err != nil {
		r.handleProfileErr(responder, err)
		return true
	}
	if person == nil {
		responder.Empty(http.StatusNotFound)
		return true
	}
	return false
}
