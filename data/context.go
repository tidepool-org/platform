package data

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
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

type Context interface {
	Logger() log.Logger

	SetMeta(meta interface{})

	ResolveReference(reference interface{}) string

	AppendError(reference interface{}, err *service.Error)

	NewChildContext(reference interface{}) Context
}
