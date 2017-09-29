package service

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
)

const (
	HTTPHeaderTraceRequest = "X-Tidepool-Trace-Request"
	HTTPHeaderTraceSession = "X-Tidepool-Trace-Session"
)

const (
	_RequestEnvErrors       = "ERRORS"
	_RequestEnvLogger       = "LOGGER"
	_RequestEnvAuthDetails  = "AUTH-DETAILS"
	_RequestEnvTraceRequest = "TRACE-REQUEST"
	_RequestEnvTraceSession = "TRACE-SESSION"
)

func GetRequestErrors(req *rest.Request) []*Error {
	if req != nil {
		if errs, ok := req.Env[_RequestEnvErrors].([]*Error); ok {
			return errs
		}
	}
	return nil
}

func SetRequestErrors(req *rest.Request, errs []*Error) {
	if req != nil {
		if errs != nil {
			req.Env[_RequestEnvErrors] = errs
		} else {
			delete(req.Env, _RequestEnvErrors)
		}
	}
}

func GetRequestLogger(req *rest.Request) log.Logger {
	if req != nil {
		if logger, ok := req.Env[_RequestEnvLogger].(log.Logger); ok {
			return logger
		}
	}
	return nil
}

func SetRequestLogger(req *rest.Request, logger log.Logger) {
	if req != nil {
		if logger != nil {
			req.Env[_RequestEnvLogger] = logger
		} else {
			delete(req.Env, _RequestEnvLogger)
		}
	}
}

func GetRequestAuthDetails(req *rest.Request) request.Details {
	if req != nil {
		if details, ok := req.Env[_RequestEnvAuthDetails].(request.Details); ok {
			return details
		}
	}
	return nil
}

func SetRequestAuthDetails(req *rest.Request, details request.Details) {
	if req != nil {
		if details != nil {
			req.Env[_RequestEnvAuthDetails] = details
		} else {
			delete(req.Env, _RequestEnvAuthDetails)
		}
	}
}

func GetRequestTraceRequest(req *rest.Request) string {
	if req != nil {
		if traceRequest, ok := req.Env[_RequestEnvTraceRequest].(string); ok {
			return traceRequest
		}
	}
	return ""
}

func SetRequestTraceRequest(req *rest.Request, traceRequest string) {
	if req != nil {
		if traceRequest != "" {
			req.Env[_RequestEnvTraceRequest] = traceRequest
		} else {
			delete(req.Env, _RequestEnvTraceRequest)
		}
	}
}

func GetRequestTraceSession(req *rest.Request) string {
	if req != nil {
		if traceSession, ok := req.Env[_RequestEnvTraceSession].(string); ok {
			return traceSession
		}
	}
	return ""
}

func SetRequestTraceSession(req *rest.Request, traceSession string) {
	if req != nil {
		if traceSession != "" {
			req.Env[_RequestEnvTraceSession] = traceSession
		} else {
			delete(req.Env, _RequestEnvTraceSession)
		}
	}
}

func CopyRequestTrace(sourceRequest *rest.Request, destinationRequest *http.Request) error {
	if sourceRequest == nil {
		return errors.New("source request is missing")
	}
	if destinationRequest == nil {
		return errors.New("destination request is missing")
	}

	if traceRequest := GetRequestTraceRequest(sourceRequest); traceRequest != "" {
		destinationRequest.Header.Add(HTTPHeaderTraceRequest, traceRequest)
	}
	if traceSession := GetRequestTraceSession(sourceRequest); traceSession != "" {
		destinationRequest.Header.Add(HTTPHeaderTraceSession, traceSession)
	}

	return nil
}
