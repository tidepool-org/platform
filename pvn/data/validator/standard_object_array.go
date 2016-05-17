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

type StandardObjectArray struct {
	context   data.Context
	reference interface{}
	value     *[]map[string]interface{}
}

func NewStandardObjectArray(context data.Context, reference interface{}, value *[]map[string]interface{}) *StandardObjectArray {
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

func (s *StandardObjectArray) LengthInRange(lowerlimit int, upperLimit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length < lowerlimit || length > upperLimit {
			s.context.AppendError(s.reference, ErrorLengthNotInRange(length, lowerlimit, upperLimit))
		}
	}
	return s
}
