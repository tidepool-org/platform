package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	platform "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/api"
)

func AlertsRoutes() []service.Route {
	return []service.Route{
		service.Get("/v1/users/:followedUserID/followers/:userID/alerts", GetAlert, api.RequireAuth),
		service.Post("/v1/users/:followedUserID/followers/:userID/alerts", UpsertAlert, api.RequireAuth),
		service.Delete("/v1/users/:followedUserID/followers/:userID/alerts", DeleteAlert, api.RequireAuth),
	}
}

func DeleteAlert(dCtx service.Context) {
	r := dCtx.Request()
	ctx := r.Context()
	authDetails := request.GetAuthDetails(ctx)
	repo := dCtx.AlertsRepository()

	if err := checkAuthentication(authDetails); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	if err := checkUserIDConsistency(authDetails, r.PathParam("userID")); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	followedUserID := r.PathParam("followedUserID")
	userID := userIDWithServiceFallback(authDetails, r.PathParam("userID"))
	pc := dCtx.PermissionClient()
	if err := checkUserAuthorization(ctx, pc, userID, followedUserID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{UserID: userID, FollowedUserID: followedUserID}
	if err := repo.Delete(ctx, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorInternalServerFailure())
		return
	}
}

func GetAlert(dCtx service.Context) {
	r := dCtx.Request()
	ctx := r.Context()
	authDetails := request.GetAuthDetails(ctx)
	repo := dCtx.AlertsRepository()

	if err := checkAuthentication(authDetails); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	followedUserID := r.PathParam("followedUserID")
	userID := userIDWithServiceFallback(authDetails, r.PathParam("userID"))
	pc := dCtx.PermissionClient()
	if err := checkUserAuthorization(ctx, pc, userID, followedUserID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	if err := checkUserIDConsistency(authDetails, r.PathParam("userID")); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{UserID: userID, FollowedUserID: followedUserID}
	alert, err := repo.Get(ctx, cfg)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			dCtx.RespondWithStatusAndErrors(http.StatusNotFound,
				[]*platform.Error{platform.ErrorValueNotExists()})
			return
		}
		dCtx.RespondWithError(platform.ErrorInternalServerFailure())
		return
	}

	responder := request.MustNewResponder(dCtx.Response(), r)
	responder.Data(http.StatusOK, alert)
}

func UpsertAlert(dCtx service.Context) {
	r := dCtx.Request()
	ctx := r.Context()
	authDetails := request.GetAuthDetails(ctx)
	repo := dCtx.AlertsRepository()
	lgr := log.LoggerFromContext(ctx)

	if err := checkAuthentication(authDetails); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	if err := checkUserIDConsistency(authDetails, r.PathParam("userID")); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	a := &alerts.Alerts{}
	if err := request.DecodeRequestBody(r.Request, a); err != nil {
		dCtx.RespondWithError(platform.ErrorJSONMalformed())
		return
	}

	followedUserID := r.PathParam("followedUserID")
	userID := userIDWithServiceFallback(authDetails, r.PathParam("userID"))
	pc := dCtx.PermissionClient()
	if err := checkUserAuthorization(ctx, pc, userID, followedUserID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{UserID: userID, FollowedUserID: followedUserID, Alerts: *a}
	if err := repo.Upsert(ctx, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorInternalServerFailure())
		lgr.WithError(err).Error("upserting alerts config")
		return
	}
}

var ErrUnauthorized = fmt.Errorf("unauthorized")

// checkUserIDConsistency verifies the userIDs in a request.
//
// For safety reasons, if these values don't agree, return an error.
func checkUserIDConsistency(details request.AuthDetails, userIDFromPath string) error {
	if details.IsService() && details.UserID() == "" {
		return nil
	}
	if details.IsUser() && userIDFromPath == details.UserID() {
		return nil
	}

	return ErrUnauthorized
}

// checkAuthentication ensures that the request has an authentication token.
func checkAuthentication(details request.AuthDetails) error {
	if details.Token() == "" {
		return ErrUnauthorized
	}
	if details.IsUser() {
		return nil
	}
	if details.IsService() {
		return nil
	}
	return ErrUnauthorized
}

// checkUserAuthorization returns nil if userID is permitted to have alerts
// based on followedUserID's data.
func checkUserAuthorization(ctx context.Context, pc permission.Client, userID, followedUserID string) error {
	perms, err := pc.GetUserPermissions(ctx, userID, followedUserID)
	if err != nil {
		return err
	}
	for key := range perms {
		if key == permission.Follow {
			return nil
		}
	}
	return fmt.Errorf("user isn't authorized for alerting: %q", userID)
}

// userIDWithServiceFallback returns the user's ID.
//
// If the request is from a user, the userID found in the token will be
// returned. This could be an empty string if the request details are
// malformed.
//
// If the request is from a service, then the service fallback value is used,
// as no userID is passed with the details in the event of a service request.
func userIDWithServiceFallback(details request.AuthDetails, serviceFallback string) string {
	if details.IsUser() {
		return details.UserID()
	}
	return serviceFallback
}
