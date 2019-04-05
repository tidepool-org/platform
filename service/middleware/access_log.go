package middleware

import (
	"fmt"
	"net"
	"time"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

type AccessLog struct{}

const (
	_LogErrors           = "errors"
	_LogMethod           = "method"
	_LogProto            = "proto"
	_LogRefererer        = "referrer"
	_LogRemoteAddress    = "remote-address"
	_LogRemoteUser       = "remote-user"
	_LogRequestURI       = "request-uri"
	_LogResponseBytes    = "response-bytes"
	_LogResponseDuration = "response-duration"
	_LogStartTime        = "start-time"
	_LogStatusCode       = "status-code"
	_LogUserAgent        = "user-agent"

	_RequestEnvBytesWritten = "BYTE_WRITTEN"
	_RequestEnvElapsedTime  = "ELAPSED_TIME"
	_RequestEnvRemoteUser   = "REMOTE_USER"
	_RequestEnvStartTime    = "START_TIME"
	_RequestEnvStatusCode   = "STATUS_CODE"
)

func NewAccessLog() (*AccessLog, error) {
	return &AccessLog{}, nil
}

func (a *AccessLog) ignore(req *rest.Request) bool {
	// ignore liveness and readiness probes
	if req.URL.RequestURI() == "/status") {
		return true
	}
	return false
}

// MiddlewareFunc adds logging of all calls, except those ignored above.
func (a *AccessLog) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handler != nil && res != nil && req != nil {
			handler(res, req)

			if a.ignore(req) {
				return
			}

			// DEPRECATED: Needs to be replaced with context version
			if logger := service.GetRequestLogger(req); logger != nil {
				loggerFields := map[string]interface{}{}
				if method := req.Method; method != "" {
					loggerFields[_LogMethod] = method
				}
				if uri := req.URL.RequestURI(); uri != "" {
					loggerFields[_LogRequestURI] = uri
				}
				if proto := req.Proto; proto != "" {
					loggerFields[_LogProto] = proto
				}
				if agent := req.UserAgent(); agent != "" {
					loggerFields[_LogUserAgent] = agent
				}
				if referer := req.Referer(); referer != "" {
					loggerFields[_LogRefererer] = referer
				}
				if remoteAddress := req.RemoteAddr; remoteAddress != "" {
					if ip, _, err := net.SplitHostPort(remoteAddress); err == nil && ip != "" {
						loggerFields[_LogRemoteAddress] = ip
					}
				}
				if remoteUser, ok := req.Env[_RequestEnvRemoteUser].(string); ok && remoteUser != "" {
					loggerFields[_LogRemoteUser] = remoteUser
				}
				if startTime, ok := req.Env[_RequestEnvStartTime].(*time.Time); ok {
					loggerFields[_LogStartTime] = startTime.Truncate(time.Microsecond).Format(time.RFC3339Nano)
				}
				if statusCode, ok := req.Env[_RequestEnvStatusCode].(int); ok {
					loggerFields[_LogStatusCode] = statusCode
				}
				if responseTime, ok := req.Env[_RequestEnvElapsedTime].(*time.Duration); ok {
					loggerFields[_LogResponseDuration] = *responseTime / time.Microsecond
				}
				if bytesWritten, ok := req.Env[_RequestEnvBytesWritten].(int64); ok {
					loggerFields[_LogResponseBytes] = bytesWritten
				}
				if errors := service.GetRequestErrors(req); errors != nil {
					loggerFields[_LogErrors] = errors // TODO: Limit to a maximum size?
				}

				message := fmt.Sprintf("%s %s", req.Method, req.URL.RequestURI())

				// TODO: Log this to a special logger level "access".
				// Will need to revamp log package, though.

				logger.WithError(request.GetErrorFromContext(req.Context())).WithFields(loggerFields).Warn(message)
			}
		}
	}
}
