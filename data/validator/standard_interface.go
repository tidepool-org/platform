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
