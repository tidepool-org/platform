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
		s.context.AppendError(s.reference, service.ErrorValueNotExists())
	}
	return s
}

func (s *StandardInteger) NotExists() data.Integer {
	if s.value != nil {
		s.context.AppendError(s.reference, service.ErrorValueExists())
	}
	return s
}

func (s *StandardInteger) EqualTo(value int) data.Integer {
	if s.value != nil {
		if *s.value != value {
			s.context.AppendError(s.reference, service.ErrorValueNotEqualTo(*s.value, value))
		}
	}
	return s
}

func (s *StandardInteger) NotEqualTo(value int) data.Integer {
	if s.value != nil {
		if *s.value == value {
			s.context.AppendError(s.reference, service.ErrorValueEqualTo(*s.value, value))
		}
	}
	return s
}

func (s *StandardInteger) LessThan(limit int) data.Integer {
	if s.value != nil {
		if *s.value >= limit {
			s.context.AppendError(s.reference, service.ErrorValueNotLessThan(*s.value, limit))
		}
	}
	return s
}

func (s *StandardInteger) LessThanOrEqualTo(limit int) data.Integer {
	if s.value != nil {
		if *s.value > limit {
			s.context.AppendError(s.reference, service.ErrorValueNotLessThanOrEqualTo(*s.value, limit))
		}
	}
	return s
}

func (s *StandardInteger) GreaterThan(limit int) data.Integer {
	if s.value != nil {
		if *s.value <= limit {
			s.context.AppendError(s.reference, service.ErrorValueNotGreaterThan(*s.value, limit))
		}
	}
	return s
}

func (s *StandardInteger) GreaterThanOrEqualTo(limit int) data.Integer {
	if s.value != nil {
		if *s.value < limit {
			s.context.AppendError(s.reference, service.ErrorValueNotGreaterThanOrEqualTo(*s.value, limit))
		}
	}
	return s
}

func (s *StandardInteger) InRange(lowerLimit int, upperLimit int) data.Integer {
	if s.value != nil {
		if *s.value < lowerLimit || *s.value > upperLimit {
			s.context.AppendError(s.reference, service.ErrorValueNotInRange(*s.value, lowerLimit, upperLimit))
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
		s.context.AppendError(s.reference, service.ErrorValueIntegerNotOneOf(*s.value, allowedValues))
	}
	return s
}

func (s *StandardInteger) NotOneOf(disallowedValues []int) data.Integer {
	if s.value != nil {
		for _, possibleValue := range disallowedValues {
			if possibleValue == *s.value {
				s.context.AppendError(s.reference, service.ErrorValueIntegerOneOf(*s.value, disallowedValues))
				return s
			}
		}
	}
	return s
}
