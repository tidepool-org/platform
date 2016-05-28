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

type StandardString struct {
	context   data.Context
	reference interface{}
	value     *string
}

func NewStandardString(context data.Context, reference interface{}, value *string) *StandardString {
	if context == nil {
		return nil
	}

	return &StandardString{
		context:   context,
		reference: reference,
		value:     value,
	}
}

func (s *StandardString) Exists() data.String {
	if s.value == nil {
		s.context.AppendError(s.reference, ErrorValueDoesNotExist())
	}
	return s
}

func (s *StandardString) EqualTo(value string) data.String {
	if s.value != nil {
		if *s.value != value {
			s.context.AppendError(s.reference, ErrorValueNotEqualTo(*s.value, value))
		}
	}
	return s
}

func (s *StandardString) NotEqualTo(value string) data.String {
	if s.value != nil {
		if *s.value == value {
			s.context.AppendError(s.reference, ErrorValueEqualTo(*s.value, value))
		}
	}
	return s
}

func (s *StandardString) LengthEqualTo(limit int) data.String {
	if s.value != nil {
		if length := len(*s.value); length != limit {
			s.context.AppendError(s.reference, ErrorLengthNotEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardString) LengthNotEqualTo(limit int) data.String {
	if s.value != nil {
		if length := len(*s.value); length == limit {
			s.context.AppendError(s.reference, ErrorLengthEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardString) LengthLessThan(limit int) data.String {
	if s.value != nil {
		if length := len(*s.value); length >= limit {
			s.context.AppendError(s.reference, ErrorLengthNotLessThan(length, limit))
		}
	}
	return s
}

func (s *StandardString) LengthLessThanOrEqualTo(limit int) data.String {
	if s.value != nil {
		if length := len(*s.value); length > limit {
			s.context.AppendError(s.reference, ErrorLengthNotLessThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardString) LengthGreaterThan(limit int) data.String {
	if s.value != nil {
		if length := len(*s.value); length <= limit {
			s.context.AppendError(s.reference, ErrorLengthNotGreaterThan(length, limit))
		}
	}
	return s
}

func (s *StandardString) LengthGreaterThanOrEqualTo(limit int) data.String {
	if s.value != nil {
		if length := len(*s.value); length < limit {
			s.context.AppendError(s.reference, ErrorLengthNotGreaterThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardString) LengthInRange(lowerLimit int, upperLimit int) data.String {
	if s.value != nil {
		if length := len(*s.value); length < lowerLimit || length > upperLimit {
			s.context.AppendError(s.reference, ErrorLengthNotInRange(length, lowerLimit, upperLimit))
		}
	}
	return s
}

func (s *StandardString) OneOf(allowedValues []string) data.String {
	if s.value != nil {
		for _, possibleValue := range allowedValues {
			if possibleValue == *s.value {
				return s
			}
		}
		s.context.AppendError(s.reference, ErrorStringNotOneOf(*s.value, allowedValues))
	}
	return s
}

func (s *StandardString) NotOneOf(disallowedValues []string) data.String {
	if s.value != nil {
		for _, possibleValue := range disallowedValues {
			if possibleValue == *s.value {
				s.context.AppendError(s.reference, ErrorStringOneOf(*s.value, disallowedValues))
				return s
			}
		}
	}
	return s
}
