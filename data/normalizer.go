package data

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

type Normalizer interface {
	Logger() log.Logger

	SetMeta(meta interface{})

	AppendError(reference interface{}, err *service.Error)

	AppendDatum(datum Datum)

	NewChildNormalizer(reference interface{}) Normalizer
}
