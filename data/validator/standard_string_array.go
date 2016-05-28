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
		s.context.AppendError(s.reference, ErrorValueDoesNotExist())
	}
	return s
}

func (s *StandardStringArray) LengthEqualTo(limit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length != limit {
			s.context.AppendError(s.reference, ErrorLengthNotEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardStringArray) LengthNotEqualTo(limit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length == limit {
			s.context.AppendError(s.reference, ErrorLengthEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardStringArray) LengthLessThan(limit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length >= limit {
			s.context.AppendError(s.reference, ErrorLengthNotLessThan(length, limit))
		}
	}
	return s
}

func (s *StandardStringArray) LengthLessThanOrEqualTo(limit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length > limit {
			s.context.AppendError(s.reference, ErrorLengthNotLessThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardStringArray) LengthGreaterThan(limit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length <= limit {
			s.context.AppendError(s.reference, ErrorLengthNotGreaterThan(length, limit))
		}
	}
	return s
}

func (s *StandardStringArray) LengthGreaterThanOrEqualTo(limit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length < limit {
			s.context.AppendError(s.reference, ErrorLengthNotGreaterThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardStringArray) LengthInRange(lowerLimit int, upperLimit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length < lowerLimit || length > upperLimit {
			s.context.AppendError(s.reference, ErrorLengthNotInRange(length, lowerLimit, upperLimit))
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
			context.AppendError(index, ErrorStringNotOneOf(value, allowedValues))
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
					context.AppendError(index, ErrorStringOneOf(value, disallowedValues))
					continue outer
				}
			}
		}
	}
	return s
}
