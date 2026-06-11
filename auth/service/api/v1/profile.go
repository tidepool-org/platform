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

		rest.Get("/users/:userId/users", r.requireCustodian("userId", r.GetUsersWithProfiles)),
		rest.Get("/metadata/users/:userId/users", r.requireCustodian("userId", r.GetUsersWithProfiles)),

		rest.Get("/v1/users/legacy/:userId/profile", r.requireMembership("userId", r.GetLegacyProfile)),
		rest.Get("/metadata/:userId/profile", r.requireMembership("userId", r.GetLegacyProfile)),

		rest.Put("/v1/users/:userId/profile", r.requireCustodian("userId", r.UpdateProfile)),
		rest.Post("/v1/users/:userId/profile", r.requireCustodian("userId", r.UpdateProfile)),

		rest.Put("/v1/users/legacy/:userId/profile", r.requireCustodian("userId", r.UpdateLegacyProfile)),
		rest.Put("/metadata/:userId/profile", r.requireCustodian("userId", r.UpdateLegacyProfile)),

		rest.Post("/v1/users/legacy/:userId/profile", r.requireCustodian("userId", r.UpdateLegacyProfile)),
		rest.Post("/metadata/:userId/profile", r.requireCustodian("userId", r.UpdateLegacyProfile)),

		rest.Delete("/v1/users/:userId/profile", r.requireCustodian("userId", r.DeleteProfile)),
		rest.Delete("/v1/users/legacy/:userId/profile", r.requireCustodian("userId", r.DeleteProfile)),
	}
}

func (r *Router) getProfile(ctx context.Context, userID string) (*user.LegacyUserProfile, error) {
	profile, err := r.ProfileAccessor().FindLegacyUserProfile(ctx, userID)
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

	responder.Data(http.StatusOK, profile.ToUserProfile())
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
	trustorPerms, err := r.PermissionsClient().PermissionsGrantedToUser(ctx, targetUserID)
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

	trusteePerms, err := r.PermissionsClient().PermissionsGrantedByUser(ctx, targetUserID)
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
	results := user.TrustUserArray{}
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
			profile := seagullProfile.ToUserProfile()
			trustUser := &user.TrustUser{
				User: *sharedUser,
				TrustPermissions: user.TrustPermissions{
					TrusteePermissions: trustPerms.TrusteePermissions,
					TrustorPermissions: trustPerms.TrustorPermissions,
				},
			}
			trustUser.Profile = profile
			lock.Lock()
			results = append(results, trustUser)
			lock.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		r.handleUserOrProfileErr(responder, err)
		return
	}

	// type TrustUserArray implements Sanitize to hide any properties for non service requests
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
	if err := r.ProfileAccessor().UpdateLegacyUserProfile(ctx, userID, profile); err != nil {
		r.handleUserOrProfileErr(responder, err)
		return
	}
	responder.Data(http.StatusOK, profile)
}

func (r *Router) UpdateProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	userID := req.PathParam("userId")

	profile := &user.Profile{}
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

func (r *Router) DeleteProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	responder.Empty(http.StatusNotImplemented)
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
