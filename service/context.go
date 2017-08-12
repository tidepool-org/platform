package service

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/log"
)

type Context interface {
	Response() rest.ResponseWriter
	Request() *rest.Request

	Logger() log.Logger

	RespondWithError(err *Error)
	RespondWithInternalServerFailure(message string, failure ...interface{})
	RespondWithStatusAndErrors(statusCode int, errors []*Error)
	RespondWithStatusAndData(statusCode int, data interface{})

	AuthDetails() auth.Details
}
