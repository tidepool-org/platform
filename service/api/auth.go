package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/user"
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
			responder := request.MustNewResponder(res, req)
			targetUserID, err := request.DecodeRequestPathParameter(req, targetParamUserID, user.IsValidID)
			if err != nil {
				responder.Error(http.StatusBadRequest, err)
				return
			}
			ctx := req.Context()
			details := request.GetAuthDetails(ctx)
			if details == nil {
				request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
				return
			}
			hasMembership, err := CheckMembership(req, permissionsClient(), targetUserID)
			if responder.RespondIfError(err) {
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

// RequireWritePermissions will proceed with the provided handlerFunc if the authenticated user has write permisisons to the userID defined in the URL param userIDParam.
//
// This will be true if the userID is one of:
//   - the same as authenticated user in AuthDetails
//   - the authenticated entity is a Service
//   - the authenticated user in AuthDetails has explicit (permissions actually defined in gatekeeper) write permissions to the userID
//
// For example:
//
//	rest.Post("/v1/myroute/:userId/action", api.RequireWritePermissions(permissionsClient, "userId", handlerFunc))
func RequireWritePermissions(permissionsClient func() permission.Client, userIDParam string, handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handlerFunc != nil && res != nil && req != nil {
			responder := request.MustNewResponder(res, req)
			targetUserID, err := request.DecodeRequestPathParameter(req, userIDParam, user.IsValidID)
			if err != nil {
				responder.Error(http.StatusBadRequest, err)
				return
			}
			ctx := req.Context()
			details := request.GetAuthDetails(ctx)
			if details == nil {
				responder.Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
				return
			}
			if !details.IsService() && details.UserID() != targetUserID {
				hasPerms, err := permission.HasExplicitWritePermissions(ctx, permissionsClient(), details.UserID(), targetUserID)
				if responder.RespondIfError(err) {
					return
				}
				if !hasPerms {
					responder.Empty(http.StatusForbidden)
					return
				}
			}
			handlerFunc(res, req)
		}
	}
}

// CheckMembership returns whether the user or service associated with the
// request has a relationship with the user whose id is targetUserID. It is not
// a middleware, but is used by some, because there are certain cases where we
// do not know the actual user id - for example, in the case of device logs, we
// have to retrieve an object and that object contains the user's id. Only
// after getting this user id we are able to check if user has permissions to
// that user.
func CheckMembership(req *rest.Request, client permission.Client, targetUserID string) (allowed bool, err error) {
	ctx := req.Context()
	details := request.GetAuthDetails(ctx)
	if details == nil {
		return false, nil
	}
	if details.IsService() || details.UserID() == targetUserID {
		return true, nil
	}
	hasPerms, err := permission.HasExplicitMembershipRelationship(ctx, client, details.UserID(), targetUserID)
	if err != nil {
		return false, err
	}
	if !hasPerms {
		return false, nil
	}
	return true, nil
}
