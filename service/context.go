package service

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

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
