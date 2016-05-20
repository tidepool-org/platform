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
}
