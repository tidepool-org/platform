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

type StandardInteger struct {
	context   data.Context
	reference interface{}
	value     *int
}

func NewStandardInteger(context data.Context, reference interface{}, value *int) *StandardInteger {
	return &StandardInteger{
		context:   context,
		reference: reference,
		value:     value,
	}
}

func (s *StandardInteger) Exists() data.Integer {
	if s.value == nil {
		s.context.AppendError(s.reference, ErrorValueDoesNotExist())
	}
	return s
}

func (s *StandardInteger) EqualTo(value int) data.Integer {
	if s.value != nil {
		if *s.value != value {
			s.context.AppendError(s.reference, ErrorValueNotEqualTo(*s.value, value))
		}
	}
	return s
}

func (s *StandardInteger) NotEqualTo(value int) data.Integer {
	if s.value != nil {
		if *s.value == value {
			s.context.AppendError(s.reference, ErrorValueEqualTo(*s.value, value))
		}
	}
	return s
}

func (s *StandardInteger) LessThan(limit int) data.Integer {
	if s.value != nil {
		if *s.value >= limit {
			s.context.AppendError(s.reference, ErrorValueNotLessThan(*s.value, limit))
		}
	}
	return s
}

func (s *StandardInteger) LessThanOrEqualTo(limit int) data.Integer {
	if s.value != nil {
		if *s.value > limit {
			s.context.AppendError(s.reference, ErrorValueNotLessThanOrEqual(*s.value, limit))
		}
	}
	return s
}

func (s *StandardInteger) GreaterThan(limit int) data.Integer {
	if s.value != nil {
		if *s.value <= limit {
			s.context.AppendError(s.reference, ErrorValueNotGreaterThan(*s.value, limit))
		}
	}
	return s
}

func (s *StandardInteger) GreaterThanOrEqualTo(limit int) data.Integer {
	if s.value != nil {
		if *s.value < limit {
			s.context.AppendError(s.reference, ErrorValueNotGreaterThanOrEqual(*s.value, limit))
		}
	}
	return s
}

func (s *StandardInteger) InRange(lowerlimit int, upperLimit int) data.Integer {
	if s.value != nil {
		if *s.value < lowerlimit || *s.value > upperLimit {
			s.context.AppendError(s.reference, ErrorIntegerNotInRange(*s.value, lowerlimit, upperLimit))
		}
	}
	return s
}

func (s *StandardInteger) OneOf(allowedValues []int) data.Integer {
	if s.value != nil {
		for _, possibleValue := range allowedValues {
			if possibleValue == *s.value {
				return s
			}
		}
		s.context.AppendError(s.reference, ErrorIntegerNotOneOf(*s.value, allowedValues))
	}
	return s
}

func (s *StandardInteger) NotOneOf(disallowedValues []int) data.Integer {
	if s.value != nil {
		for _, possibleValue := range disallowedValues {
			if possibleValue == *s.value {
				s.context.AppendError(s.reference, ErrorIntegerOneOf(*s.value, disallowedValues))
				return s
			}
		}
	}
	return s
}
