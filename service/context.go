package service

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/log"
)

type Context interface {
	Logger() log.Logger

	Request() *rest.Request
	Response() rest.ResponseWriter

	RespondWithError(err *Error)
	RespondWithInternalServerFailure(message string, failure ...interface{})
	RespondWithStatusAndErrors(statusCode int, errors []*Error)
	RespondWithStatusAndData(statusCode int, data interface{})
}
