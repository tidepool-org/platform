package service

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

const (
	HTTPHeaderTraceRequest = "X-Tidepool-Trace-Request"
	HTTPHeaderTraceSession = "X-Tidepool-Trace-Session"
)

const (
	_RequestEnvErrors       = "ERRORS"
	_RequestEnvLogger       = "LOGGER"
	_RequestEnvTraceRequest = "TRACE-REQUEST"
	_RequestEnvTraceSession = "TRACE-SESSION"
)

func GetRequestErrors(request *rest.Request) []*Error {
	if request != nil {
		if errors, ok := request.Env[_RequestEnvErrors].([]*Error); ok {
			return errors
		}
	}
	return nil
}

func SetRequestErrors(request *rest.Request, errors []*Error) {
	if request != nil {
		if errors != nil {
			request.Env[_RequestEnvErrors] = errors
		} else {
			delete(request.Env, _RequestEnvErrors)
		}
	}
}

func GetRequestLogger(request *rest.Request) log.Logger {
	if request != nil {
		if logger, ok := request.Env[_RequestEnvLogger].(log.Logger); ok {
			return logger
		}
	}
	return nil
}

func SetRequestLogger(request *rest.Request, logger log.Logger) {
	if request != nil {
		if logger != nil {
			request.Env[_RequestEnvLogger] = logger
		} else {
			delete(request.Env, _RequestEnvLogger)
		}
	}
}

func GetRequestTraceRequest(request *rest.Request) string {
	if request != nil {
		if traceRequest, ok := request.Env[_RequestEnvTraceRequest].(string); ok {
			return traceRequest
		}
	}
	return ""
}

func SetRequestTraceRequest(request *rest.Request, traceRequest string) {
	if request != nil {
		if traceRequest != "" {
			request.Env[_RequestEnvTraceRequest] = traceRequest
		} else {
			delete(request.Env, _RequestEnvTraceRequest)
		}
	}
}

func GetRequestTraceSession(request *rest.Request) string {
	if request != nil {
		if traceSession, ok := request.Env[_RequestEnvTraceSession].(string); ok {
			return traceSession
		}
	}
	return ""
}

func SetRequestTraceSession(request *rest.Request, traceSession string) {
	if request != nil {
		if traceSession != "" {
			request.Env[_RequestEnvTraceSession] = traceSession
		} else {
			delete(request.Env, _RequestEnvTraceSession)
		}
	}
}

func CopyRequestTrace(sourceRequest *rest.Request, destinationRequest *http.Request) error {
	if sourceRequest == nil {
		return errors.New("service", "source request is missing")
	}
	if destinationRequest == nil {
		return errors.New("service", "destination request is missing")
	}

	if traceRequest := GetRequestTraceRequest(sourceRequest); traceRequest != "" {
		destinationRequest.Header.Add(HTTPHeaderTraceRequest, traceRequest)
	}
	if traceSession := GetRequestTraceSession(sourceRequest); traceSession != "" {
		destinationRequest.Header.Add(HTTPHeaderTraceSession, traceSession)
	}

	return nil
}
