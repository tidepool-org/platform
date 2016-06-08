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

type StandardInterfaceArray struct {
	context   data.Context
	reference interface{}
	value     *[]interface{}
}

func NewStandardInterfaceArray(context data.Context, reference interface{}, value *[]interface{}) *StandardInterfaceArray {
	if context == nil {
		return nil
	}

	return &StandardInterfaceArray{
		context:   context,
		reference: reference,
		value:     value,
	}
}

func (s *StandardInterfaceArray) Exists() data.InterfaceArray {
	if s.value == nil {
		s.context.AppendError(s.reference, ErrorValueNotExists())
	}
	return s
}

func (s *StandardInterfaceArray) NotExists() data.InterfaceArray {
	if s.value != nil {
		s.context.AppendError(s.reference, ErrorValueExists())
	}
	return s
}

func (s *StandardInterfaceArray) Empty() data.InterfaceArray {
	if s.value != nil {
		if len(*s.value) != 0 {
			s.context.AppendError(s.reference, ErrorValueNotEmpty())
		}
	}
	return s
}

func (s *StandardInterfaceArray) NotEmpty() data.InterfaceArray {
	if s.value != nil {
		if len(*s.value) == 0 {
			s.context.AppendError(s.reference, ErrorValueEmpty())
		}
	}
	return s
}

func (s *StandardInterfaceArray) LengthEqualTo(limit int) data.InterfaceArray {
	if s.value != nil {
		if length := len(*s.value); length != limit {
			s.context.AppendError(s.reference, ErrorLengthNotEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardInterfaceArray) LengthNotEqualTo(limit int) data.InterfaceArray {
	if s.value != nil {
		if length := len(*s.value); length == limit {
			s.context.AppendError(s.reference, ErrorLengthEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardInterfaceArray) LengthLessThan(limit int) data.InterfaceArray {
	if s.value != nil {
		if length := len(*s.value); length >= limit {
			s.context.AppendError(s.reference, ErrorLengthNotLessThan(length, limit))
		}
	}
	return s
}

func (s *StandardInterfaceArray) LengthLessThanOrEqualTo(limit int) data.InterfaceArray {
	if s.value != nil {
		if length := len(*s.value); length > limit {
			s.context.AppendError(s.reference, ErrorLengthNotLessThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardInterfaceArray) LengthGreaterThan(limit int) data.InterfaceArray {
	if s.value != nil {
		if length := len(*s.value); length <= limit {
			s.context.AppendError(s.reference, ErrorLengthNotGreaterThan(length, limit))
		}
	}
	return s
}

func (s *StandardInterfaceArray) LengthGreaterThanOrEqualTo(limit int) data.InterfaceArray {
	if s.value != nil {
		if length := len(*s.value); length < limit {
			s.context.AppendError(s.reference, ErrorLengthNotGreaterThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardInterfaceArray) LengthInRange(lowerLimit int, upperLimit int) data.InterfaceArray {
	if s.value != nil {
		if length := len(*s.value); length < lowerLimit || length > upperLimit {
			s.context.AppendError(s.reference, ErrorLengthNotInRange(length, lowerLimit, upperLimit))
		}
	}
	return s
}
