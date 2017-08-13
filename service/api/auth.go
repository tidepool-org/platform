package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/context"
)

func AuthServer(handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		if handlerFunc != nil && response != nil && request != nil {
			ctx, err := context.New(response, request)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				return
			}

			if authDetails := ctx.AuthDetails(); authDetails == nil || !authDetails.IsServer() {
				ctx.RespondWithError(service.ErrorUnauthenticated())
				return
			}

			handlerFunc(response, request)
		}
	}
}

func AuthUser(handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		if handlerFunc != nil && response != nil && request != nil {
			ctx, err := context.New(response, request)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				return
			}

			if authDetails := ctx.AuthDetails(); authDetails == nil {
				ctx.RespondWithError(service.ErrorUnauthenticated())
				return
			}

			handlerFunc(response, request)
		}
	}
}
