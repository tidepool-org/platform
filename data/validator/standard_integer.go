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

type StandardInteger struct {
	context   data.Context
	reference interface{}
	value     *int
}

func NewStandardInteger(context data.Context, reference interface{}, value *int) *StandardInteger {
	if context == nil {
		return nil
	}

	return &StandardInteger{
		context:   context,
		reference: reference,
		value:     value,
	}
}

func (s *StandardInteger) Exists() data.Integer {
	if s.value == nil {
		s.context.AppendError(s.reference, ErrorValueNotExists())
	}
	return s
}

func (s *StandardInteger) NotExists() data.Integer {
	if s.value != nil {
		s.context.AppendError(s.reference, ErrorValueExists())
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
			s.context.AppendError(s.reference, ErrorValueNotLessThanOrEqualTo(*s.value, limit))
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
			s.context.AppendError(s.reference, ErrorValueNotGreaterThanOrEqualTo(*s.value, limit))
		}
	}
	return s
}

func (s *StandardInteger) InRange(lowerLimit int, upperLimit int) data.Integer {
	if s.value != nil {
		if *s.value < lowerLimit || *s.value > upperLimit {
			s.context.AppendError(s.reference, ErrorIntegerNotInRange(*s.value, lowerLimit, upperLimit))
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
