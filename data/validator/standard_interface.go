package validator

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/service"
)

type StandardInterface struct {
	context   data.Context
	reference interface{}
	value     *interface{}
}

func NewStandardInterface(context data.Context, reference interface{}, value *interface{}) *StandardInterface {
	if context == nil {
		return nil
	}

	return &StandardInterface{
		context:   context,
		reference: reference,
		value:     value,
	}
}

func (s *StandardInterface) Exists() data.Interface {
	if s.value == nil {
		s.context.AppendError(s.reference, service.ErrorValueNotExists())
	}
	return s
}

func (s *StandardInterface) NotExists() data.Interface {
	if s.value != nil {
		s.context.AppendError(s.reference, service.ErrorValueExists())
	}
	return s
}
