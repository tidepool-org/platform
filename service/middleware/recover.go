package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
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
						stack := strings.Split(strings.ReplaceAll(string(debug.Stack()), "\t", ""), "\n")
						err := errors.WithMeta(errors.Newf("unhandled panic: %v", r), stack)
						responder.RespondWithInternalServerFailure("Recovered from unhandled panic", err)
					}
				}
			}()

			handler(res, req)
		}
	}
}
