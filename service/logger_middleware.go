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
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
)

type LoggerMiddleware struct {
	Logger log.Logger
}

const (
	RequestEnvLogger = "logger"
)

func NewLoggerMiddleware(logger log.Logger) (*LoggerMiddleware, error) {
	if logger == nil {
		return nil, app.Error("service", "logger is missing")
	}

	return &LoggerMiddleware{logger}, nil
}

func (l *LoggerMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		request.Env[RequestEnvLogger] = l.Logger
		handler(response, request)
	}
}

func GetRequestLogger(request *rest.Request) log.Logger {
	if request != nil {
		if logger, ok := request.Env[RequestEnvLogger].(log.Logger); ok {
			return logger
		}
	}
	return nil // TODO: Should probably return something other than nil here
}
