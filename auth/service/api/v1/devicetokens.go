package v1

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
)

func (r *Router) DeviceTokensRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/users/:userId/device_tokens", api.RequireAuth(r.UpsertDeviceToken)),
	}
}

func (r *Router) UpsertDeviceToken(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Request.Context()
	authDetails := request.GetAuthDetails(ctx)
	repo := r.AuthStore().NewDeviceTokenRepository()

	if err := checkAuthentication(authDetails); err != nil {
		log.Printf("checkAuth failed: %+v", authDetails)
		responder.Error(http.StatusUnauthorized, err)
		return
	}

	if err := checkUserIDConsistency(authDetails, req.PathParam("userId")); err != nil {
		log.Printf("checkUserIDConsistency failed: %+v %q", authDetails, req.PathParam("userID"))
		responder.Error(http.StatusForbidden, err)
		return
	}

	deviceToken := devicetokens.DeviceToken{}
	if err := request.DecodeRequestBody(req.Request, &deviceToken); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	userID := userIDWithServiceFallback(authDetails, req.PathParam("userId"))
	doc := devicetokens.NewDocument(userID, deviceToken)
	if err := repo.Upsert(ctx, doc); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}
}

var ErrUnauthorized = fmt.Errorf("unauthorized")

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
