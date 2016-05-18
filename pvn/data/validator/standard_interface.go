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
		s.context.AppendError(s.reference, ErrorValueDoesNotExist())
	}
	return s
}
