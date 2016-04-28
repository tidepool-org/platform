package service

import (
	"fmt"
	"net"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
)

type AccessLogMiddleware struct{}

const (
	LogErrors           = "errors"
	LogMethod           = "method"
	LogProto            = "proto"
	LogRefererer        = "referrer"
	LogRemoteAddress    = "remote-address"
	LogRemoteUser       = "remote-user"
	LogRequestURI       = "request-uri"
	LogResponseBytes    = "response-bytes"
	LogResponseDuration = "response-duration"
	LogStartTime        = "start-time"
	LogStatusCode       = "status-code"
	LogUserAgent        = "user-agent"

	RequestEnvBytesWritten = "BYTE_WRITTEN"
	RequestEnvElapsedTime  = "ELAPSED_TIME"
	RequestEnvErrors       = "ERRORS"
	RequestEnvRemoteUser   = "REMOTE_USER"
	RequestEnvStartTime    = "START_TIME"
	RequestEnvStatusCode   = "STATUS_CODE"
)

func (l *AccessLogMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		handler(response, request)

		loggerFields := map[string]interface{}{}
		if method := request.Method; method != "" {
			loggerFields[LogMethod] = method
		}
		if uri := request.URL.RequestURI(); uri != "" {
			loggerFields[LogRequestURI] = uri
		}
		if proto := request.Proto; proto != "" {
			loggerFields[LogProto] = proto
		}
		if agent := request.UserAgent(); agent != "" {
			loggerFields[LogUserAgent] = agent
		}
		if referer := request.Referer(); referer != "" {
			loggerFields[LogRefererer] = referer
		}
		if remoteAddress := request.RemoteAddr; remoteAddress != "" {
			if ip, _, err := net.SplitHostPort(remoteAddress); err == nil && ip != "" {
				loggerFields[LogRemoteAddress] = ip
			}
		}
		if remoteUser, ok := request.Env[RequestEnvRemoteUser].(string); ok && remoteUser != "" {
			loggerFields[LogRemoteUser] = remoteUser
		}
		if startTime, ok := request.Env[RequestEnvStartTime].(*time.Time); ok {
			loggerFields[LogStartTime] = startTime.Format(time.RFC3339)
		}
		if statusCode, ok := request.Env[RequestEnvStatusCode].(int); ok {
			loggerFields[LogStatusCode] = statusCode
		}
		if responseTime, ok := request.Env[RequestEnvElapsedTime].(*time.Duration); ok {
			loggerFields[LogResponseDuration] = *responseTime / time.Microsecond
		}
		if bytesWritten, ok := request.Env[RequestEnvBytesWritten].(int64); ok {
			loggerFields[LogResponseBytes] = bytesWritten
		}
		if errors := request.Env[RequestEnvErrors]; errors != nil {
			loggerFields[LogErrors] = errors // TODO: Limit to a maximum size?
		}

		message := fmt.Sprintf("%s %s", request.Method, request.URL.RequestURI())

		GetRequestLogger(request).WithFields(loggerFields).Info(message)
	}
}
