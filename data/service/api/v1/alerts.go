package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

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

	if err := checkAuthentication(details, r.PathParam("userID")); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{}
	if err := request.DecodeRequestBody(r.Request, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorJSONMalformed())
	}

	pc := dCtx.PermissionClient()
	if err := checkUserAuthorization(ctx, pc, details.UserID(), cfg.FollowedID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg.UserID = details.UserID()
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

	if err := checkAuthentication(details, r.PathParam("userID")); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	followedID := r.PathParam("followedID")
	pc := dCtx.PermissionClient()
	if err := checkUserAuthorization(ctx, pc, details.UserID(), followedID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{UserID: details.UserID(), FollowedID: followedID}
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

	if err := checkAuthentication(details, r.PathParam("userID")); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{}
	if err := json.NewDecoder(r.Body).Decode(cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorJSONMalformed())
	}

	pc := dCtx.PermissionClient()
	if err := checkUserAuthorization(ctx, pc, details.UserID(), cfg.FollowedID); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg.UserID = details.UserID()
	if err := repo.Upsert(ctx, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorInternalServerFailure())
		return
	}
}

var ErrUnauthorized = fmt.Errorf("unauthorized")

func checkAuthentication(details request.Details, userID string) error {
	if details.IsUser() {
		if details.UserID() != userID {
			log.Printf("warning: URL userID doesn't match token userID, token wins ")
		}
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
