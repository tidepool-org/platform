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

type StandardObjectArray struct {
	context   data.Context
	reference interface{}
	value     *[]map[string]interface{}
}

func NewStandardObjectArray(context data.Context, reference interface{}, value *[]map[string]interface{}) *StandardObjectArray {
	if context == nil {
		return nil
	}

	return &StandardObjectArray{
		context:   context,
		reference: reference,
		value:     value,
	}
}

func (s *StandardObjectArray) Exists() data.ObjectArray {
	if s.value == nil {
		s.context.AppendError(s.reference, ErrorValueDoesNotExist())
	}
	return s
}

func (s *StandardObjectArray) NotEmpty() data.ObjectArray {
	if s.value != nil {
		if len(*s.value) == 0 {
			s.context.AppendError(s.reference, ErrorValueEmpty())
		}
	}
	return s
}

func (s *StandardObjectArray) LengthEqualTo(limit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length != limit {
			s.context.AppendError(s.reference, ErrorLengthNotEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardObjectArray) LengthNotEqualTo(limit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length == limit {
			s.context.AppendError(s.reference, ErrorLengthEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardObjectArray) LengthLessThan(limit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length >= limit {
			s.context.AppendError(s.reference, ErrorLengthNotLessThan(length, limit))
		}
	}
	return s
}

func (s *StandardObjectArray) LengthLessThanOrEqualTo(limit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length > limit {
			s.context.AppendError(s.reference, ErrorLengthNotLessThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardObjectArray) LengthGreaterThan(limit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length <= limit {
			s.context.AppendError(s.reference, ErrorLengthNotGreaterThan(length, limit))
		}
	}
	return s
}

func (s *StandardObjectArray) LengthGreaterThanOrEqualTo(limit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length < limit {
			s.context.AppendError(s.reference, ErrorLengthNotGreaterThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardObjectArray) LengthInRange(lowerLimit int, upperLimit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length < lowerLimit || length > upperLimit {
			s.context.AppendError(s.reference, ErrorLengthNotInRange(length, lowerLimit, upperLimit))
		}
	}
	return s
}
