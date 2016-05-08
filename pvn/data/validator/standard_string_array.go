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

type StandardStringArray struct {
	context   data.Context
	reference interface{}
	value     *[]string
}

func NewStandardStringArray(context data.Context, reference interface{}, value *[]string) *StandardStringArray {
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
			s.context.AppendError(s.reference, ErrorLengthNotLessThanOrEqual(length, limit))
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
			s.context.AppendError(s.reference, ErrorLengthNotGreaterThanOrEqual(length, limit))
		}
	}
	return s
}

func (s *StandardStringArray) LengthInRange(lowerlimit int, upperLimit int) data.StringArray {
	if s.value != nil {
		if length := len(*s.value); length < lowerlimit || length > upperLimit {
			s.context.AppendError(s.reference, ErrorLengthNotInRange(length, lowerlimit, upperLimit))
		}
	}
	return s
}

func (s *StandardStringArray) EachOneOf(possibleValues []string) data.StringArray {
	if s.value != nil {
		context := s.context.NewChildContext(s.reference)
	outer:
		for index, value := range *s.value {
			for _, possibleValue := range possibleValues {
				if possibleValue == value {
					continue outer
				}
			}
			context.AppendError(index, ErrorStringNotOneOf(value, possibleValues))
		}
	}
	return s
}
