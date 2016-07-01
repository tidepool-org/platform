package middleware

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"runtime/debug"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/service/context"
)

type Recover struct{}

func NewRecover() (*Recover, error) {
	return &Recover{}, nil
}

func (r *Recover) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		if handler != nil && response != nil && request != nil {
			defer func() {
				if r := recover(); r != nil {
					standardContext, _ := context.NewStandard(response, request)
					if standardContext != nil {
						standardContext.RespondWithInternalServerFailure("Recovered from unhandled panic", string(debug.Stack()))
					}
				}
			}()

			handler(response, request)
		}
	}
}
