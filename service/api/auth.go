package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
)

// RequireAuth aborts with an error if a request isn't authenticated.
//
// Requests with incorrect, invalid, or no credentials are rejected.
//
// RequireAuth works by checking for the existence of an AuthDetails sentinel
// value, which implies an important assumption:
//
//	The existence of AuthDetails in the request's Context indicates that the
//	request has already been properly authenticated.
//
// The function that performs the actual authentication is in the
// service/middleware package. As long as no other code adds an AuthDetails
// value to the request's Context (when the request isn't properly
// authenticated) then things should be good.
func RequireAuth(handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handlerFunc != nil && res != nil && req != nil {
			if details := request.GetAuthDetails(req.Context()); details == nil {
				request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
			} else {
				handlerFunc(res, req)
			}
		}
	}
}

// RequireServer aborts with an error if a request isn't authenticated as a server.
func RequireServer(handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handlerFunc != nil && res != nil && req != nil {
			if details := request.GetAuthDetails(req.Context()); details == nil {
				request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
			} else if !details.IsService() {
				request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
			} else {
				handlerFunc(res, req)
			}
		}
	}
}

// RequireUser aborts with an error if a request isn't authenticated as a user.
func RequireUser(handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handlerFunc != nil && res != nil && req != nil {
			if details := request.GetAuthDetails(req.Context()); details == nil {
				request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
			} else if !details.IsUser() {
				request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
			} else {
				handlerFunc(res, req)
			}
		}
	}
}

// RequireMembership aborts a handler if there is no connection between the
// user with the id identified in the url param targetParamUserID and the
// request is not a service to service request. It takes a function that
// returns a permission.Client instead of one directly because this may be used
// within middleware where the client has not been created yet.
func RequireMembership(permissionsClient func() permission.Client, targetParamUserID string, handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handlerFunc != nil && res != nil && req != nil {
			targetUserID := req.PathParam(targetParamUserID)
			responder := request.MustNewResponder(res, req)
			ctx := req.Context()
			details := request.GetAuthDetails(ctx)
			if details == nil {
				request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
				return
			}
			hasMembership, err := CheckMembership(req, permissionsClient(), targetUserID)
			if err != nil {
				responder.InternalServerError(err)
				return
			}
			if !hasMembership {
				responder.Empty(http.StatusForbidden)
				return
			}
			handlerFunc(res, req)
		}
	}
}

func CheckMembership(req *rest.Request, client permission.Client, targetUserID string) (allowed bool, err error) {
	ctx := req.Context()
	details := request.GetAuthDetails(ctx)
	if details == nil {
		return false, nil
	}
	if details.IsService() || details.UserID() == targetUserID {
		return true, nil
	}
	hasPerms, err := permission.HasMembershipRelationship(ctx, client, details.UserID(), targetUserID)
	if err != nil {
		return false, err
	}
	if !hasPerms {
		return false, nil
	}
	return true, nil
}
