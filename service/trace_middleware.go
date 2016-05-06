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
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/app"
)

type TraceMiddleware struct{}

const (
	HTTPHeaderTraceRequest = "X-Tidepool-Trace-Request"
	HTTPHeaderTraceSession = "X-Tidepool-Trace-Session"

	LogTraceRequest = "trace-request"
	LogTraceSession = "trace-session"

	RequestEnvTraceRequest = "trace-request"
	RequestEnvTraceSession = "trace-session"

	TraceSessionMaximumLength = 64
)

func NewTraceMiddleware() (*TraceMiddleware, error) {
	return &TraceMiddleware{}, nil
}

func (l *TraceMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		logger := GetRequestLogger(request)

		traceRequest := app.NewUUID()
		request.Env[RequestEnvTraceRequest] = traceRequest
		response.Header().Add(HTTPHeaderTraceRequest, traceRequest)
		logger = logger.WithField(LogTraceRequest, traceRequest)

		if traceSession, ok := request.Header[HTTPHeaderTraceSession]; ok {
			if len(traceSession) > TraceSessionMaximumLength {
				traceSession = traceSession[:TraceSessionMaximumLength]
			}
			request.Env[RequestEnvTraceSession] = traceSession
			logger = logger.WithField(LogTraceSession, traceSession)
		}

		request.Env[RequestEnvLogger] = logger

		handler(response, request)
	}
}

func GetRequestTraceRequest(request *rest.Request) string {
	if request != nil {
		if traceRequest, ok := request.Env[RequestEnvTraceRequest].(string); ok {
			return traceRequest
		}
	}
	return ""
}

func GetRequestTraceSession(request *rest.Request) string {
	if request != nil {
		if traceSession, ok := request.Env[RequestEnvTraceSession].(string); ok {
			return traceSession
		}
	}
	return ""
}

func CopyRequestTrace(sourceRequest *rest.Request, destinationRequest *http.Request) error {
	if sourceRequest == nil {
		return app.Error("service", "source request is missing")
	}
	if destinationRequest == nil {
		return app.Error("service", "destination request is missing")
	}

	if traceRequest := GetRequestTraceRequest(sourceRequest); traceRequest != "" {
		destinationRequest.Header.Add(HTTPHeaderTraceRequest, traceRequest)
	}
	if traceSession := GetRequestTraceSession(sourceRequest); traceSession != "" {
		destinationRequest.Header.Add(HTTPHeaderTraceSession, traceSession)
	}

	return nil
}
