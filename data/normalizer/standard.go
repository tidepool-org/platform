package normalizer

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/log"
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

func (s *Standard) Logger() log.Logger {
	return s.context.Logger()
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

func (s *Standard) NewChildNormalizer(reference interface{}) data.Normalizer {
	return &Standard{
		context: s.context.NewChildContext(reference),
		data:    s.data,
	}
}
