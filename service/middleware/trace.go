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
	"github.com/tidepool-org/platform/service"
)

type Trace struct{}

const (
	_LogTraceRequest = "trace-request"
	_LogTraceSession = "trace-session"

	_TraceSessionMaximumLength = 64
)

func NewTrace() (*Trace, error) {
	return &Trace{}, nil
}

func (l *Trace) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		if handler != nil && response != nil && request != nil {
			oldLogger := service.GetRequestLogger(request)
			oldTraceRequest := service.GetRequestTraceRequest(request)
			oldTraceSession := service.GetRequestTraceSession(request)

			defer func() {
				service.SetRequestTraceSession(request, oldTraceSession)
				service.SetRequestTraceRequest(request, oldTraceRequest)
				service.SetRequestLogger(request, oldLogger)
			}()

			newLogger := oldLogger

			newTraceRequest := app.NewID()
			service.SetRequestTraceRequest(request, newTraceRequest)
			if newLogger != nil {
				newLogger = newLogger.WithField(_LogTraceRequest, newTraceRequest)
			}

			if newTraceSession := request.Header.Get(service.HTTPHeaderTraceSession); newTraceSession != "" {
				if len(newTraceSession) > _TraceSessionMaximumLength {
					newTraceSession = newTraceSession[:_TraceSessionMaximumLength]
				}
				service.SetRequestTraceSession(request, newTraceSession)
				if newLogger != nil {
					newLogger = newLogger.WithField(_LogTraceSession, newTraceSession)
				}
			}

			service.SetRequestLogger(request, newLogger)

			response.Header().Add(service.HTTPHeaderTraceRequest, newTraceRequest)

			handler(response, request)
		}
	}
}
