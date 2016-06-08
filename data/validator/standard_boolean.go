package validator

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import "github.com/tidepool-org/platform/data"

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
		s.context.AppendError(s.reference, ErrorValueNotExists())
	}
	return s
}

func (s *StandardBoolean) NotExists() data.Boolean {
	if s.value != nil {
		s.context.AppendError(s.reference, ErrorValueExists())
	}
	return s
}

func (s *StandardBoolean) True() data.Boolean {
	if s.value != nil {
		if !*s.value {
			s.context.AppendError(s.reference, ErrorValueNotTrue())
		}
	}
	return s
}

func (s *StandardBoolean) False() data.Boolean {
	if s.value != nil {
		if *s.value {
			s.context.AppendError(s.reference, ErrorValueNotFalse())
		}
	}
	return s
}
