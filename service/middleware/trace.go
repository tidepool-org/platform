package middleware

import (
	"github.com/mdblp/go-json-rest/rest"

	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

type Trace struct{}

const (
	_LogTrace   = "trace"
	_LogRequest = "request"
	_LogSession = "session"

	_TraceMaximumLength = 64
)

func NewTrace() (*Trace, error) {
	return &Trace{}, nil
}

func (t *Trace) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handler != nil && res != nil && req != nil {
			oldRequest := req.Request
			defer func() {
				req.Request = oldRequest
			}()

			trace := map[string]interface{}{}

			// DEPRECATED
			oldTraceRequest := service.GetRequestTraceRequest(req)
			defer service.SetRequestTraceRequest(req, oldTraceRequest)

			traceRequest := req.Header.Get(request.HTTPHeaderTraceRequest)
			if traceRequest != "" {
				if len(traceRequest) > _TraceMaximumLength {
					traceRequest = traceRequest[:_TraceMaximumLength]
				}
			} else {
				traceRequest = id.Must(id.New(16))
			}
			req.Request = req.WithContext(request.NewContextWithTraceRequest(req.Context(), traceRequest))
			service.SetRequestTraceRequest(req, traceRequest) // DEPRECATED
			res.Header().Add(request.HTTPHeaderTraceRequest, traceRequest)
			trace[_LogRequest] = traceRequest

			traceSession := req.Header.Get(request.HTTPHeaderTraceSession)
			if traceSession != "" {
				// DEPRECATED
				oldTraceSession := service.GetRequestTraceSession(req)
				defer service.SetRequestTraceSession(req, oldTraceSession)

				if len(traceSession) > _TraceMaximumLength {
					traceSession = traceSession[:_TraceMaximumLength]
				}
				req.Request = req.WithContext(request.NewContextWithTraceSession(req.Context(), traceSession))
				service.SetRequestTraceSession(req, traceSession) // DEPRECATED
				res.Header().Add(request.HTTPHeaderTraceSession, traceSession)
				trace[_LogSession] = traceSession
			}

			// DEPRECATED
			if oldLogger := service.GetRequestLogger(req); oldLogger != nil {
				defer service.SetRequestLogger(req, oldLogger)
				service.SetRequestLogger(req, oldLogger.WithField(_LogTrace, trace))
			}

			if logger := log.LoggerFromContext(req.Context()); logger != nil {
				req.Request = req.WithContext(log.NewContextWithLogger(req.Context(), logger.WithField(_LogTrace, trace)))
			}

			handler(res, req)
		}
	}
}
