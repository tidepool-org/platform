package middleware

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

type Logger struct {
	logger log.Logger
}

func NewLogger(logger log.Logger) (*Logger, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}

	return &Logger{
		logger: logger,
	}, nil
}

func (l *Logger) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handler != nil && res != nil && req != nil {
			oldRequest := req.Request
			defer func() {
				req.Request = oldRequest
			}()

			// DEPRECATED
			oldLogger := service.GetRequestLogger(req)
			defer service.SetRequestLogger(req, oldLogger)
			service.SetRequestLogger(req, l.logger)

			req.Request = req.WithContext(log.NewContextWithLogger(req.Context(), l.logger))

			handler(res, req)
		}
	}
}
