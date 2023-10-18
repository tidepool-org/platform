package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

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
