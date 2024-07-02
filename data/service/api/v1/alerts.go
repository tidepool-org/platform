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
	platformerrors "github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	platform "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/api"
)

func AlertsRoutes() []service.Route {
	return []service.Route{
		service.Get("/v1/users/:userId/followers/:followerUserId/alerts", GetAlert, api.RequireAuth),
		service.Post("/v1/users/:userId/followers/:followerUserId/alerts", UpsertAlert, api.RequireAuth),
		service.Delete("/v1/users/:userId/followers/:followerUserId/alerts", DeleteAlert, api.RequireAuth),
		service.Get("/v1/users/:userId/followers/alerts", ListAlerts, api.RequireServer),
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

	path := getUserIDsFromPath(r)
	if err := checkUserIDConsistency(authDetails, path.UserID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	pc := dCtx.PermissionClient()
	if err := checkUserAuthorization(ctx, pc, userIDWithServiceFallback(authDetails, path.UserID), path.FollowedUserID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{UserID: path.UserID, FollowedUserID: path.FollowedUserID}
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

	path := getUserIDsFromPath(r)
	pc := dCtx.PermissionClient()
	if err := checkUserAuthorization(ctx, pc, userIDWithServiceFallback(authDetails, path.UserID), path.FollowedUserID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	if err := checkUserIDConsistency(authDetails, path.UserID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{UserID: path.UserID, FollowedUserID: path.FollowedUserID}
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

	path := getUserIDsFromPath(r)
	if err := checkUserIDConsistency(authDetails, path.UserID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	incomingCfg := &alerts.Config{}
	var bodyReceiver interface{} = &incomingCfg.Alerts
	if authDetails.IsService() && authDetails.UserID() == "" {
		// Accept upload id only from services.
		bodyReceiver = incomingCfg
	}
	if err := request.DecodeRequestBody(r.Request, bodyReceiver); err != nil {
		dCtx.RespondWithError(platform.ErrorJSONMalformed())
		return
	}

	pc := dCtx.PermissionClient()
	if err := checkUserAuthorization(ctx, pc, userIDWithServiceFallback(authDetails, path.UserID), path.FollowedUserID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{
		UserID:         path.UserID,
		FollowedUserID: path.FollowedUserID,
		UploadID:       incomingCfg.UploadID,
		Alerts:         incomingCfg.Alerts,
	}
	if err := repo.Upsert(ctx, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorInternalServerFailure())
		lgr.WithError(err).Error("upserting alerts config")
		return
	}
}

func ListAlerts(dCtx service.Context) {
	r := dCtx.Request()
	ctx := r.Context()
	authDetails := request.GetAuthDetails(ctx)
	repo := dCtx.AlertsRepository()
	lgr := log.LoggerFromContext(ctx)

	if err := checkAuthentication(authDetails); err != nil {
		lgr.Debug("authentication failed")
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	pathsUserID := r.PathParam("userId")
	if err := checkUserIDConsistency(authDetails, pathsUserID); err != nil {
		lgr.WithFields(log.Fields{"path": pathsUserID, "auth": authDetails.UserID()}).
			Debug("user id consistency failed")
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	alerts, err := repo.List(ctx, pathsUserID)
	if err != nil {
		dCtx.RespondWithInternalServerFailure("listing alerts configs", err)
		lgr.WithError(err).Error("listing alerts config")
		return
	}
	if len(alerts) == 0 {
		dCtx.RespondWithError(ErrorUserIDNotFound(pathsUserID))
		lgr.Debug("no alerts configs found")
	}

	responder := request.MustNewResponder(dCtx.Response(), r)
	responder.Data(http.StatusOK, alerts)
}

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

	return platformerrors.New("unauthorized")
}

// checkAuthentication ensures that the request has an authentication token.
func checkAuthentication(details request.AuthDetails) error {
	if details.Token() == "" {
		return platformerrors.New("unauthorized")
	}
	if details.IsUser() {
		return nil
	}
	if details.IsService() {
		return nil
	}
	return platformerrors.New("unauthorized")
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

// alertsUserIDs prevents confusion about the roles of the user ids found in alerts endpoint
// paths.
type alertsUserIDs struct {
	FollowedUserID string
	UserID         string
}

func getUserIDsFromPath(r *rest.Request) alertsUserIDs {
	// Within alerts endpoints handlers, the following user owns the alerts config, so are
	// called "userID". Due to restrictions in the github.com/ant0ine/go-json-rest library, the
	// path parameters can't be renamed to better reflect this situation.
	return alertsUserIDs{
		FollowedUserID: r.PathParam("userId"),
		UserID:         r.PathParam("followerUserId"),
	}
}
