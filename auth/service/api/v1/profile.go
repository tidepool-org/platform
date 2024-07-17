package v1

import (
	"context"
	stdErrs "errors"
	"maps"
	"net/http"
	"sync"

	"github.com/ant0ine/go-json-rest/rest"
	"golang.org/x/sync/errgroup"

	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	structValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

type trustPermissions struct {
	TrustorPermissions *permission.Permission
	TrusteePermissions *permission.Permission
}

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

	mergedUserPerms := map[string]*trustPermissions{}
	trustorPerms, err := r.PermissionsClient().GroupsForUser(ctx, targetUserID)
	if err != nil {
		responder.InternalServerError(err)
		return
	}
	for userID, perms := range trustorPerms {
		if userID == targetUserID {
			// Don't include own user in result
			continue
		}

		clone := maps.Clone(perms)
		mergedUserPerms[userID] = &trustPermissions{
			TrustorPermissions: &clone,
		}
	}

	trusteePerms, err := r.PermissionsClient().UsersInGroup(ctx, targetUserID)
	if err != nil {
		responder.InternalServerError(err)
		return
	}
	for userID, perms := range trusteePerms {
		if userID == targetUserID {
			// Don't include own user in result
			continue
		}

		if _, ok := mergedUserPerms[userID]; !ok {
			mergedUserPerms[userID] = &trustPermissions{}
		}
		clone := maps.Clone(perms)
		mergedUserPerms[userID].TrusteePermissions = &clone
	}

	lock := &sync.Mutex{}
	results := make([]*user.User, 0, len(mergedUserPerms))
	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(20) // do up to 20 concurrent requests like seagull did
	for userID, trustPerms := range mergedUserPerms {
		userID, trustPerms := userID, trustPerms
		group.Go(func() error {
			sharedUser, err := r.UserAccessor().FindUserById(ctx, userID)
			if stdErrs.Is(err, user.ErrUserNotFound) || sharedUser == nil {
				// According to seagull code, "It's possible for a user profile to be deleted before the sharing permissions", so we can ignore if user or profile not found.
				return nil
			}
			if err != nil {
				return err
			}
			profile, err := r.getProfile(ctx, userID)
			if stdErrs.Is(err, user.ErrUserProfileNotFound) || profile == nil {
				return nil
			}
			if err != nil {
				return err
			}
			trustorPerms := trustPerms.TrustorPermissions
			if trustorPerms == nil || len(*trustorPerms) == 0 {
				profile = profile.ClearPatientInfo()
			} else {
				if trustorPerms.HasAny(permission.Custodian, permission.Read, permission.Write) {
					// TODO: need to read seagull.value.settings - confirm this is actually used
				}
				if trustorPerms.Has(permission.Custodian) {
					// TODO: need to read seagull.value.preferences - confirm this is actually used
				}
			}
			sharedUser.Profile = profile
			sharedUser.TrusteePermissions = trustPerms.TrusteePermissions
			sharedUser.TrustorPermissions = trustPerms.TrustorPermissions
			// Seems no sharedUser.Sanitize call to filter out "protected" fields in seagull except sanitizeUser to remove "passwordExists" field - which doesn't exist in current platform/user.User
			lock.Lock()
			results = append(results, sharedUser)
			lock.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		r.handleProfileErr(responder, err)
		return
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
