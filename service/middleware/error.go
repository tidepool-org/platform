package middleware

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/request"
)

type Error struct{}

func NewError() (*Error, error) {
	return &Error{}, nil
}

func (e *Error) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handler != nil && res != nil && req != nil {
			oldRequest := req.Request
			defer func() {
				req.Request = oldRequest
			}()
			req.Request = req.WithContext(request.NewContextWithContextError(req.Context()))

			handler(res, req)
		}
	}
}
