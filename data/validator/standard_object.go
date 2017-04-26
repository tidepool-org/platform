package validator

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/service"
)

type StandardObject struct {
	context   data.Context
	reference interface{}
	value     *map[string]interface{}
}

func NewStandardObject(context data.Context, reference interface{}, value *map[string]interface{}) *StandardObject {
	if context == nil {
		return nil
	}

	return &StandardObject{
		context:   context,
		reference: reference,
		value:     value,
	}
}

func (s *StandardObject) Exists() data.Object {
	if s.value == nil {
		s.context.AppendError(s.reference, service.ErrorValueNotExists())
	}
	return s
}

func (s *StandardObject) NotExists() data.Object {
	if s.value != nil {
		s.context.AppendError(s.reference, service.ErrorValueExists())
	}
	return s
}

func (s *StandardObject) Empty() data.Object {
	if s.value != nil {
		if len(*s.value) != 0 {
			s.context.AppendError(s.reference, service.ErrorValueNotEmpty())
		}
	}
	return s
}

func (s *StandardObject) NotEmpty() data.Object {
	if s.value != nil {
		if len(*s.value) == 0 {
			s.context.AppendError(s.reference, service.ErrorValueEmpty())
		}
	}
	return s
}
