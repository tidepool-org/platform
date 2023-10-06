package v1

import (
	data "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/request"
	platform "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/api"
)

func DeviceTokensRoutes() []data.Route {
	return []data.Route{
		data.MakeRoute("POST", "/v1/device-tokens/:userID", UpsertDeviceToken, api.RequireAuth),
	}
}

func UpsertDeviceToken(dCtx data.Context) {
	r := dCtx.Request()
	ctx := r.Context()
	authDetails := request.GetAuthDetails(ctx)
	repo := dCtx.DeviceTokensRepository()

	if err := checkAuthentication(authDetails); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	if err := checkUserIDConsistency(authDetails, r.PathParam("userID")); err != nil {
		dCtx.RespondWithError(platform.ErrorUnauthorized())
		return
	}

	deviceToken := devicetokens.DeviceToken{}
	if err := request.DecodeRequestBody(r.Request, &deviceToken); err != nil {
		dCtx.RespondWithError(platform.ErrorJSONMalformed())
		return
	}

	userID := userIDWithServiceFallback(authDetails, r.PathParam("userID"))
	doc := devicetokens.NewDocument(userID, deviceToken)
	if err := repo.Upsert(ctx, doc); err != nil {
		dCtx.RespondWithError(platform.ErrorInternalServerFailure())
		return
	}
}
