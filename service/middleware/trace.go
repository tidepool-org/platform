package middleware

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/service"
)

type Trace struct{}

const (
	_LogTraceRequest = "trace-request"
	_LogTraceSession = "trace-session"

	_TraceMaximumLength = 64
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

			newTraceRequest := request.Header.Get(service.HTTPHeaderTraceRequest)
			if newTraceRequest != "" {
				if len(newTraceRequest) > _TraceMaximumLength {
					newTraceRequest = newTraceRequest[:_TraceMaximumLength]
				}
			} else {
				newTraceRequest = app.NewID()
			}
			service.SetRequestTraceRequest(request, newTraceRequest)
			if newLogger != nil {
				newLogger = newLogger.WithField(_LogTraceRequest, newTraceRequest)
			}
			response.Header().Add(service.HTTPHeaderTraceRequest, newTraceRequest)

			newTraceSession := request.Header.Get(service.HTTPHeaderTraceSession)
			if newTraceSession != "" {
				if len(newTraceSession) > _TraceMaximumLength {
					newTraceSession = newTraceSession[:_TraceMaximumLength]
				}
				service.SetRequestTraceSession(request, newTraceSession)
				if newLogger != nil {
					newLogger = newLogger.WithField(_LogTraceSession, newTraceSession)
				}
				response.Header().Add(service.HTTPHeaderTraceSession, newTraceSession)
			}

			service.SetRequestLogger(request, newLogger)

			handler(response, request)
		}
	}
}
