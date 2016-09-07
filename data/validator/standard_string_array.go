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

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/service"
)

type StandardStringArray struct {
	context   data.Context
	reference interface{}
	value     *[]string
}

func NewStandardStringArray(context data.Context, reference interface{}, value *[]string) *StandardStringArray {
	if context == nil {
		return nil
	}

	return &StandardStringArray{
		context:   context,
		reference: reference,
		value:     value,
	}
}

func (s *StandardStringArray) Exists() data.StringArray {
	if s.value == nil {
		s.context.AppendError(s.reference, service.ErrorValueNotExists())
	}
	return s
}

func (s *StandardStringArray) NotExists() data.StringArray {
	if s.value != nil {
		s.context.AppendError(s.reference, service.ErrorValueExists())
	}
	return s
}

func (s *StandardStringArray) Empty() data.StringArray {
	if s.value != nil {
		if len(*s.value) != 0 {
			s.context.AppendError(s.reference, service.ErrorValueNotEmpty())
		}
	}
	return s
}

func (s *StandardStringArray) NotEmpty() data.StringArray {
	if s.value != nil {
		if len(*s.value) == 0 {
			s.context.AppendError(s.reference, service.ErrorValueEmpty())
		}
	}
	return s
}

func (s *StandardStringArray) LengthEqualTo(limit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length != limit {
			s.context.AppendError(s.reference, service.ErrorLengthNotEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardStringArray) LengthNotEqualTo(limit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length == limit {
			s.context.AppendError(s.reference, service.ErrorLengthEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardStringArray) LengthLessThan(limit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length >= limit {
			s.context.AppendError(s.reference, service.ErrorLengthNotLessThan(length, limit))
		}
	}
	return s
}

func (s *StandardStringArray) LengthLessThanOrEqualTo(limit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length > limit {
			s.context.AppendError(s.reference, service.ErrorLengthNotLessThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardStringArray) LengthGreaterThan(limit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length <= limit {
			s.context.AppendError(s.reference, service.ErrorLengthNotGreaterThan(length, limit))
		}
	}
	return s
}

func (s *StandardStringArray) LengthGreaterThanOrEqualTo(limit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length < limit {
			s.context.AppendError(s.reference, service.ErrorLengthNotGreaterThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardStringArray) LengthInRange(lowerLimit int, upperLimit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length < lowerLimit || length > upperLimit {
			s.context.AppendError(s.reference, service.ErrorLengthNotInRange(length, lowerLimit, upperLimit))
		}
	}
	return s
}

func (s *StandardStringArray) EachOneOf(allowedValues []string) data.StringArray {
	if s.value != nil {
		context := s.context.NewChildContext(s.reference)
	outer:
		for index, value := range *s.value {
			for _, possibleValue := range allowedValues {
				if possibleValue == value {
					continue outer
				}
			}
			context.AppendError(index, service.ErrorValueStringNotOneOf(value, allowedValues))
		}
	}
	return s
}

func (s *StandardStringArray) EachNotOneOf(disallowedValues []string) data.StringArray {
	if s.value != nil {
		context := s.context.NewChildContext(s.reference)
	outer:
		for index, value := range *s.value {
			for _, possibleValue := range disallowedValues {
				if possibleValue == value {
					context.AppendError(index, service.ErrorValueStringOneOf(value, disallowedValues))
					continue outer
				}
			}
		}
	}
	return s
}
