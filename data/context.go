package data

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
