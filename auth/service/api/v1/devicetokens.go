package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
)

func (r *Router) DeviceTokensRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/users/:userId/device_tokens", api.RequireUser(r.UpsertDeviceToken)),
	}
}

func (r *Router) UpsertDeviceToken(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Request.Context()
	authDetails := request.GetAuthDetails(ctx)
	repo := r.AuthStore().NewDeviceTokenRepository()

	if req.PathParam("userId") != authDetails.UserID() {
		responder.Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}

	deviceToken := devicetokens.DeviceToken{}
	if err := request.DecodeRequestBody(req.Request, &deviceToken); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	doc := devicetokens.NewDocument(authDetails.UserID(), deviceToken)
	if err := repo.Upsert(ctx, doc); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}
}
