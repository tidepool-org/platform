package normalizer

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
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/service"
)

type Standard struct {
	context data.Context
	data    *[]data.Datum
}

func NewStandard(context data.Context) (*Standard, error) {
	if context == nil {
		return nil, app.Error("normalizer", "context is missing")
	}

	return &Standard{
		context: context,
		data:    &[]data.Datum{},
	}, nil
}

func (s *Standard) SetMeta(meta interface{}) {
	s.context.SetMeta(meta)
}

func (s *Standard) AppendError(reference interface{}, err *service.Error) {
	s.context.AppendError(reference, err)
}

func (s *Standard) Data() []data.Datum {
	return *s.data
}

func (s *Standard) AppendDatum(datum data.Datum) {
	if datum != nil {
		*s.data = append(*s.data, datum)
	}
}

func (s *Standard) NormalizeBloodGlucose(units *string) data.BloodGlucoseNormalizer {
	return NewStandardBloodGlucose(units)
}

func (s *Standard) NewChildNormalizer(reference interface{}) data.Normalizer {
	return &Standard{
		context: s.context.NewChildContext(reference),
		data:    s.data,
	}
}
