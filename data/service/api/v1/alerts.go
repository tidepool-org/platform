package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	platform "github.com/tidepool-org/platform/service"
)

func AlertsRoutes() []service.Route {
	return []service.Route{
		service.MakeRoute("GET", "/v1/alerts/:userID/:followedID", Authenticate(GetAlert)),
		service.MakeRoute("POST", "/v1/alerts/:userID", Authenticate(UpsertAlert)),
		service.MakeRoute("DELETE", "/v1/alerts/:userID", Authenticate(DeleteAlert)),
	}
}

func DeleteAlert(dCtx service.Context) {
	r := dCtx.Request()
	ctx := r.Context()
	details := request.DetailsFromContext(ctx)
	repo := dCtx.AlertsRepository()

	if err := checkAuthentication(details); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{}
	if err := request.DecodeRequestBody(r.Request, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorJSONMalformed())
		return
	}

	if err := checkUserIDConsistency(details, r, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorJSONMalformed())
		return
	}

	userID := userIDWithServiceFallback(details, r.PathParam("userID"))
	pc := dCtx.PermissionClient()
	if err := checkUserAuthorization(ctx, pc, userID, cfg.FollowedID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	if err := repo.Delete(ctx, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorInternalServerFailure())
		return
	}
}

func GetAlert(dCtx service.Context) {
	r := dCtx.Request()
	ctx := r.Context()
	details := request.DetailsFromContext(ctx)
	repo := dCtx.AlertsRepository()

	if err := checkAuthentication(details); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	userID := userIDWithServiceFallback(details, r.PathParam("userID"))
	followedID := r.PathParam("followedID")
	pc := dCtx.PermissionClient()
	if err := checkUserAuthorization(ctx, pc, userID, followedID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{UserID: userID, FollowedID: followedID}
	if err := checkUserIDConsistency(details, r, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorJSONMalformed())
		return
	}

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
	details := request.DetailsFromContext(ctx)
	repo := dCtx.AlertsRepository()

	if err := checkAuthentication(details); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{}
	if err := request.DecodeRequestBody(r.Request, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorJSONMalformed())
		return
	}

	if err := checkUserIDConsistency(details, r, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorJSONMalformed())
		return
	}

	userID := userIDWithServiceFallback(details, r.PathParam("userID"))
	pc := dCtx.PermissionClient()
	if err := checkUserAuthorization(ctx, pc, userID, cfg.FollowedID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	if err := repo.Upsert(ctx, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorInternalServerFailure())
		return
	}
}

var (
	ErrBadRequest   = fmt.Errorf("bad request")
	ErrUnauthorized = fmt.Errorf("unauthorized")
)

// checkUserIDConsistency verifies the various userIDs in a request.
//
// There are three possible sources of userIDs:
//  1. the request path
//  2. the alerts.Config specified in the request body
//  3. the authenticating token (if a user token)
//
// For safety reasons, if any of these three values don't agree, return an
// error (bad request).
func checkUserIDConsistency(details request.Details, r *rest.Request, cfg *alerts.Config) error {
	if details.IsService() {
		// Services won't have a userID in their token, so that check is
		// skipped.
		if r.PathParam("userID") == cfg.UserID {
			return nil
		}
		return ErrBadRequest
	}

	if r.PathParam("userID") == details.UserID() && details.UserID() == cfg.UserID {
		return nil
	}

	return ErrBadRequest
}

// checkAuthentication ensures that the request has an authentication token.
func checkAuthentication(details request.Details) error {
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
// based on followedID's data.
func checkUserAuthorization(ctx context.Context, pc permission.Client, userID, followedID string) error {
	perms, err := pc.GetUserPermissions(ctx, userID, followedID)
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
func userIDWithServiceFallback(details request.Details, serviceFallback string) string {
	if details.IsUser() {
		return details.UserID()
	}
	return serviceFallback
}
