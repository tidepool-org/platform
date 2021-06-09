package v1

import (
	"encoding/json"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
)

type VerifyTokenRequest struct {
	Token string `json:"device_token"`
}

type VerifyTokenResponse struct {
	Valid bool `json:"valid"`
}

func (r *Router) DeviceCheckRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/device_check/verify", api.RequireUser(r.VerifyToken)),
	}
}

func (r *Router) VerifyToken(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	var verifyTokenRequest VerifyTokenRequest
	err := json.NewDecoder(req.Body).Decode(&verifyTokenRequest)
	if err != nil || verifyTokenRequest.Token == "" {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	var verifyTokenResponse VerifyTokenResponse
	verifyTokenResponse.Valid, err = r.DeviceCheck().IsTokenValid(verifyTokenRequest.Token)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, verifyTokenResponse)
	return
}
