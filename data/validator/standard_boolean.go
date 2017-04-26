package validator

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/service"
)

type StandardBoolean struct {
	context   data.Context
	reference interface{}
	value     *bool
}

func NewStandardBoolean(context data.Context, reference interface{}, value *bool) *StandardBoolean {
	if context == nil {
		return nil
	}

	return &StandardBoolean{
		context:   context,
		reference: reference,
		value:     value,
	}
}

func (s *StandardBoolean) Exists() data.Boolean {
	if s.value == nil {
		s.context.AppendError(s.reference, service.ErrorValueNotExists())
	}
	return s
}

func (s *StandardBoolean) NotExists() data.Boolean {
	if s.value != nil {
		s.context.AppendError(s.reference, service.ErrorValueExists())
	}
	return s
}

func (s *StandardBoolean) True() data.Boolean {
	if s.value != nil {
		if !*s.value {
			s.context.AppendError(s.reference, service.ErrorValueNotTrue())
		}
	}
	return s
}

func (s *StandardBoolean) False() data.Boolean {
	if s.value != nil {
		if *s.value {
			s.context.AppendError(s.reference, service.ErrorValueNotFalse())
		}
	}
	return s
}
