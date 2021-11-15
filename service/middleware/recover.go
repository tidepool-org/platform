package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/mdblp/go-json-rest/rest"

	serviceContext "github.com/tidepool-org/platform/service/context"
)

type Recover struct{}

func NewRecover() (*Recover, error) {
	return &Recover{}, nil
}

func (r *Recover) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handler != nil && res != nil && req != nil {
			defer func() {
				if r := recover(); r != nil {
					if responder, responderErr := serviceContext.NewResponder(res, req); responderErr != nil {
						res.WriteHeader(http.StatusInternalServerError)
					} else {
						responder.RespondWithInternalServerFailure("Recovered from unhandled panic", string(debug.Stack()))
					}
				}
			}()

			handler(res, req)
		}
	}
}
