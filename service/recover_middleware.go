package service

import (
	"runtime/debug"

	"github.com/ant0ine/go-json-rest/rest"
)

type RecoverMiddleware struct{}

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
