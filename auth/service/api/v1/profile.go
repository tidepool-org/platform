package v1

import (
	"context"
	stdErrs "errors"
	"maps"
	"net/http"
	"sync"

	"github.com/ant0ine/go-json-rest/rest"
	"golang.org/x/sync/errgroup"

	"github.com/tidepool-org/platform/log"
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

func (r *Router) getProfile(ctx context.Context, userID string) (*user.LegacyUserProfile, error) {
	profile, err := r.ProfileAccessor().FindUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, user.ErrUserProfileNotFound
	}
	return profile, nil
}

// GetProfile returns the user's profile in the new, non seagull, format
func (r *Router) GetProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	userID := req.PathParam("userId")
	profile, err := r.getProfile(ctx, userID)
	if err != nil {
		r.handleUserOrProfileErr(responder, err)
		return
	}

	responder.Data(http.StatusOK, profile)
}

func (r *Router) GetUsersWithProfiles(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	targetUserID := req.PathParam("userId")
	targetUser, err := r.UserAccessor().FindUserById(ctx, targetUserID)
	if err != nil {
		r.handleUserOrProfileErr(responder, err)
		return
	}
	if targetUser == nil {
		r.handleUserOrProfileErr(responder, user.ErrUserNotFound)
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
	results := make(user.UserArray, 0, len(mergedUserPerms))
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
			seagullProfile, err := r.getProfile(ctx, userID)
			if stdErrs.Is(err, user.ErrUserProfileNotFound) || seagullProfile == nil {
				return nil
			}
			if err != nil {
				return err
			}
			trustorPerms := trustPerms.TrustorPermissions

			// TODO: get actual roles
			profile := seagullProfile.ToUserProfile(nil)
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
			// type UsersArray implements Sanitize to hide any properties for non service requests
			lock.Lock()
			results = append(results, sharedUser)
			lock.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		r.handleUserOrProfileErr(responder, err)
		return
	}

	responder.Data(http.StatusOK, results)
}

// GetLegacyProfile returns user profiles in the legacy seagull format.
func (r *Router) GetLegacyProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	userID := req.PathParam("userId")
	profile, err := r.getProfile(ctx, userID)
	if err != nil {
		r.handleUserOrProfileErr(responder, err)
		return
	}

	responder.Data(http.StatusOK, profile)
}

func (r *Router) UpdateLegacyProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	userID := req.PathParam("userId")

	profile := &user.LegacyUserProfile{}
	if err := request.DecodeRequestBody(req.Request, profile); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	if err := structValidator.New(log.LoggerFromContext(ctx)).Validate(profile); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	if err := r.ProfileAccessor().UpdateUserProfile(ctx, userID, profile); err != nil {
		r.handleUserOrProfileErr(responder, err)
		return
	}
	responder.Data(http.StatusOK, profile)
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
	if err := structValidator.New(log.LoggerFromContext(ctx)).Validate(profile); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	if err := r.ProfileAccessor().UpdateUserProfileV2(ctx, userID, profile); err != nil {
		r.handleUserOrProfileErr(responder, err)
		return
	}
	responder.Data(http.StatusOK, profile)
}

func (r *Router) DeleteProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	userID := req.PathParam("userId")

	err := r.ProfileAccessor().DeleteUserProfile(ctx, userID)
	if err != nil {
		r.handleUserOrProfileErr(responder, err)
		return
	}
	responder.Empty(http.StatusOK)
}

func (r *Router) handleUserOrProfileErr(responder *request.Responder, err error) {
	switch {
	case stdErrs.Is(err, user.ErrUserNotFound), stdErrs.Is(err, user.ErrUserProfileNotFound):
		// Many of the seagull clients don't treat 404 as an error so return 404 as is
		responder.Empty(http.StatusNotFound)
		return
	default:
		responder.InternalServerError(err)
	}
}

func (r *Router) handledUserNotExists(ctx context.Context, responder *request.Responder, userID string) (handled bool) {
	person, err := r.UserAccessor().FindUserById(ctx, userID)
	if err != nil {
		r.handleUserOrProfileErr(responder, err)
		return true
	}
	if person == nil {
		responder.Empty(http.StatusNotFound)
		return true
	}
	return false
}
