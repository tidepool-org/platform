package v1

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/request"
	platform "github.com/tidepool-org/platform/service"
)

func AlertsRoutes() []service.Route {
	return []service.Route{
		service.MakeRoute("POST", "/v1/alerts/:userID", Authenticate(UpsertAlert)),
		service.MakeRoute("DELETE", "/v1/alerts/:userID", Authenticate(DeleteAlert)),
	}
}

func DeleteAlert(dCtx service.Context) {
	r := dCtx.Request()
	ctx := r.Context()
	details := request.DetailsFromContext(ctx)
	repo := dCtx.AlertsRepository()

	if err := checkPermissions(details, r.PathParam("userID")); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{}
	if err := request.DecodeRequestBody(r.Request, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorJSONMalformed())
	}

	cfg.OwnerID = details.UserID()
	if err := repo.Delete(ctx, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorInternalServerFailure())
		return
	}
}

func UpsertAlert(dCtx service.Context) {
	r := dCtx.Request()
	ctx := r.Context()
	details := request.DetailsFromContext(ctx)
	repo := dCtx.AlertsRepository()

	if err := checkPermissions(details, r.PathParam("userID")); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	cfg := &alerts.Config{}
	if err := json.NewDecoder(r.Body).Decode(cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorJSONMalformed())
	}
	cfg.OwnerID = details.UserID()
	if err := repo.Upsert(ctx, cfg); err != nil {
		dCtx.RespondWithError(platform.ErrorInternalServerFailure())
		return
	}
}

var ErrUnauthorized = fmt.Errorf("unauthorized")

func checkPermissions(details request.Details, userID string) error {
	if details.IsUser() {
		if details.UserID() != userID {
			log.Printf("warning: URL userID doesn't match token userID, token wins ")
		}
		return nil
	}
	return ErrUnauthorized
}