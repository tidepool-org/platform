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
