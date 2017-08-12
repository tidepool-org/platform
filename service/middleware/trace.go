package middleware

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
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

func (t *Trace) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
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

			newFields := log.Fields{}

			newTraceRequest := request.Header.Get(service.HTTPHeaderTraceRequest)
			if newTraceRequest != "" {
				if len(newTraceRequest) > _TraceMaximumLength {
					newTraceRequest = newTraceRequest[:_TraceMaximumLength]
				}
			} else {
				newTraceRequest = id.New()
			}
			service.SetRequestTraceRequest(request, newTraceRequest)
			response.Header().Add(service.HTTPHeaderTraceRequest, newTraceRequest)
			newFields[_LogTraceRequest] = newTraceRequest

			newTraceSession := request.Header.Get(service.HTTPHeaderTraceSession)
			if newTraceSession != "" {
				if len(newTraceSession) > _TraceMaximumLength {
					newTraceSession = newTraceSession[:_TraceMaximumLength]
				}
				service.SetRequestTraceSession(request, newTraceSession)
				response.Header().Add(service.HTTPHeaderTraceSession, newTraceSession)
				newFields[_LogTraceSession] = newTraceSession
			}

			if oldLogger != nil {
				service.SetRequestLogger(request, oldLogger.WithFields(newFields))
			}

			handler(response, request)
		}
	}
}
