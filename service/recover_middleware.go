package service

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"runtime/debug"

	"github.com/ant0ine/go-json-rest/rest"
)

type RecoverMiddleware struct{}

func NewRecoverMiddleware() (*RecoverMiddleware, error) {
	return &RecoverMiddleware{}, nil
}

func (r *RecoverMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		defer func() {
			if r := recover(); r != nil {
				NewContext(response, request).RespondWithServerFailure("Recovered from unhandled panic", string(debug.Stack()))
			}
		}()

		handler(response, request)
	}
}
