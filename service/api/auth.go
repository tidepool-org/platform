package api

import (
	"net/http"

	"github.com/mdblp/go-json-rest/rest"

	"github.com/tidepool-org/platform/request"
)

func Require(handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handlerFunc != nil && res != nil && req != nil {
			if details := request.DetailsFromContext(req.Context()); details == nil {
				request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
			} else {
				handlerFunc(res, req)
			}
		}
	}
}

func RequireServer(handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handlerFunc != nil && res != nil && req != nil {
			if details := request.DetailsFromContext(req.Context()); details == nil {
				request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
			} else if !details.IsService() {
				request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
			} else {
				handlerFunc(res, req)
			}
		}
	}
}

func RequireUser(handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handlerFunc != nil && res != nil && req != nil {
			if details := request.DetailsFromContext(req.Context()); details == nil {
				request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
			} else if !details.IsUser() {
				request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
			} else {
				handlerFunc(res, req)
			}
		}
	}
}
