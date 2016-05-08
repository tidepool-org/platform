package validator

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import "github.com/tidepool-org/platform/pvn/data"

type StandardBoolean struct {
	context   data.Context
	reference interface{}
	value     *bool
}

func NewStandardBoolean(context data.Context, reference interface{}, value *bool) *StandardBoolean {
	return &StandardBoolean{
		context:   context,
		reference: reference,
		value:     value,
	}
}

func (s *StandardBoolean) Exists() data.Boolean {
	if s.value == nil {
		s.context.AppendError(s.reference, ErrorValueDoesNotExist())
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
