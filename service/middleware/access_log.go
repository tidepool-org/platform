package middleware

import (
	"fmt"
	"net"
	"time"

	"github.com/ant0ine/go-json-rest/rest"

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

func (l *AccessLog) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		if handler != nil && response != nil && request != nil {
			handler(response, request)

			if logger := service.GetRequestLogger(request); logger != nil {
				loggerFields := map[string]interface{}{}
				if method := request.Method; method != "" {
					loggerFields[_LogMethod] = method
				}
				if uri := request.URL.RequestURI(); uri != "" {
					loggerFields[_LogRequestURI] = uri
				}
				if proto := request.Proto; proto != "" {
					loggerFields[_LogProto] = proto
				}
				if agent := request.UserAgent(); agent != "" {
					loggerFields[_LogUserAgent] = agent
				}
				if referer := request.Referer(); referer != "" {
					loggerFields[_LogRefererer] = referer
				}
				if remoteAddress := request.RemoteAddr; remoteAddress != "" {
					if ip, _, err := net.SplitHostPort(remoteAddress); err == nil && ip != "" {
						loggerFields[_LogRemoteAddress] = ip
					}
				}
				if remoteUser, ok := request.Env[_RequestEnvRemoteUser].(string); ok && remoteUser != "" {
					loggerFields[_LogRemoteUser] = remoteUser
				}
				if startTime, ok := request.Env[_RequestEnvStartTime].(*time.Time); ok {
					loggerFields[_LogStartTime] = startTime.Format(time.RFC3339)
				}
				if statusCode, ok := request.Env[_RequestEnvStatusCode].(int); ok {
					loggerFields[_LogStatusCode] = statusCode
				}
				if responseTime, ok := request.Env[_RequestEnvElapsedTime].(*time.Duration); ok {
					loggerFields[_LogResponseDuration] = *responseTime / time.Microsecond
				}
				if bytesWritten, ok := request.Env[_RequestEnvBytesWritten].(int64); ok {
					loggerFields[_LogResponseBytes] = bytesWritten
				}
				if errors := service.GetRequestErrors(request); errors != nil {
					loggerFields[_LogErrors] = errors // TODO: Limit to a maximum size?
				}

				message := fmt.Sprintf("%s %s", request.Method, request.URL.RequestURI())

				// TODO: Log this to a special logger level "access".
				// Will need to revamp log package, though.

				logger.WithFields(loggerFields).Warn(message)
			}
		}
	}
}
