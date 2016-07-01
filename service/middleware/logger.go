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
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

type Logger struct {
	logger log.Logger
}

func NewLogger(logger log.Logger) (*Logger, error) {
	if logger == nil {
		return nil, app.Error("middleware", "logger is missing")
	}

	return &Logger{
		logger: logger,
	}, nil
}

func (l *Logger) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		if handler != nil && response != nil && request != nil {
			oldLogger := service.GetRequestLogger(request)

			defer func() {
				service.SetRequestLogger(request, oldLogger)
			}()

			service.SetRequestLogger(request, l.logger)

			handler(response, request)
		}
	}
}
