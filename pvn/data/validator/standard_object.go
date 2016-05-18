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
		s.context.AppendError(s.reference, ErrorValueDoesNotExist())
	}
	return s
}
