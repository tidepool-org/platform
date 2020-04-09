package v1

import (
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
)

const TidepoolBearerTokenHeaderKey = "x-tidepool-bearer-token"

func (r *Router) DeviceAuthorizationRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/users/:userId/device_authorizations", api.Require(r.ListDeviceAuthorizations)),
		rest.Post("/v1/users/:userId/device_authorizations", api.Require(r.CreateDeviceAuthorization)),
		rest.Get("/v1/users/:userId/device_authorizations/:deviceAuthorizationId", api.Require(r.GetDeviceAuthorization)),
		rest.Post("/v1/device_authorizations", r.UpdateDeviceAuthorization),
	}
}

func (r *Router) CreateDeviceAuthorization(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.DetailsFromContext(req.Context())

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	if !details.IsService() && details.UserID() != userID {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	create := auth.NewDeviceAuthorizationCreate()
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	deviceAuthorization, err := r.AuthClient().CreateUserDeviceAuthorization(req.Context(), userID, create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	res.Header().Add(TidepoolBearerTokenHeaderKey, deviceAuthorization.Token)
	responder.Data(http.StatusCreated, deviceAuthorization)
}

func (r *Router) GetDeviceAuthorization(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	details := request.DetailsFromContext(req.Context())

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	if !details.IsService() && details.UserID() != userID {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	deviceAuthorizationID := req.PathParam("deviceAuthorizationId")
	if deviceAuthorizationID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("deviceAuthorizationId"))
		return
	}

	deviceAuthorization, err := r.AuthClient().GetUserDeviceAuthorization(req.Context(), userID, deviceAuthorizationID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	if deviceAuthorization == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFound())
		return
	}

	responder.Data(http.StatusOK, deviceAuthorization)
}

func (r *Router) ListDeviceAuthorizations(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)
	details := request.DetailsFromContext(ctx)

	userID := req.PathParam("userId")
	if userID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("userId"))
		return
	}

	if !details.IsService() && details.UserID() != userID {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	deviceAuthorizations, err := r.AuthClient().ListUserDeviceAuthorizations(ctx, userID, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, deviceAuthorizations)
}

func (r *Router) UpdateDeviceAuthorization(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)

	token := getBearerToken(req)
	if token == "" {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	deviceAuthorization, err := r.AuthClient().GetDeviceAuthorizationByToken(ctx, token)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	if deviceAuthorization == nil {
		responder.Error(http.StatusUnauthorized, request.ErrorUnauthorized())
		return
	}

	update := auth.NewDeviceAuthorizationUpdate()
	if err := request.DecodeRequestBody(req.Request, update); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	if deviceAuthorization.ShouldExpire() {
		update.Expire()
	}

	updated, err := r.AuthClient().UpdateDeviceAuthorization(ctx, deviceAuthorization.ID, update)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
	}

	responder.Data(http.StatusOK, updated)
}

func getBearerToken(req *rest.Request) string {
	authzHeader := req.Header.Get("Authorization")
	bearerTokenPrefix := "Bearer "
	if !strings.HasPrefix(authzHeader, bearerTokenPrefix) {
		return ""
	}

	return strings.TrimPrefix(authzHeader, bearerTokenPrefix)
}
