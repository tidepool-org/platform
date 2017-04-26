package validator

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/service"
)

type StandardFloat struct {
	context   data.Context
	reference interface{}
	value     *float64
}

func NewStandardFloat(context data.Context, reference interface{}, value *float64) *StandardFloat {
	if context == nil {
		return nil
	}

	return &StandardFloat{
		context:   context,
		reference: reference,
		value:     value,
	}
}

func (s *StandardFloat) Exists() data.Float {
	if s.value == nil {
		s.context.AppendError(s.reference, service.ErrorValueNotExists())
	}
	return s
}

func (s *StandardFloat) NotExists() data.Float {
	if s.value != nil {
		s.context.AppendError(s.reference, service.ErrorValueExists())
	}
	return s
}

func (s *StandardFloat) EqualTo(value float64) data.Float {
	if s.value != nil {
		if *s.value != value {
			s.context.AppendError(s.reference, service.ErrorValueNotEqualTo(*s.value, value))
		}
	}
	return s
}

func (s *StandardFloat) NotEqualTo(value float64) data.Float {
	if s.value != nil {
		if *s.value == value {
			s.context.AppendError(s.reference, service.ErrorValueEqualTo(*s.value, value))
		}
	}
	return s
}

func (s *StandardFloat) LessThan(limit float64) data.Float {
	if s.value != nil {
		if *s.value >= limit {
			s.context.AppendError(s.reference, service.ErrorValueNotLessThan(*s.value, limit))
		}
	}
	return s
}

func (s *StandardFloat) LessThanOrEqualTo(limit float64) data.Float {
	if s.value != nil {
		if *s.value > limit {
			s.context.AppendError(s.reference, service.ErrorValueNotLessThanOrEqualTo(*s.value, limit))
		}
	}
	return s
}

func (s *StandardFloat) GreaterThan(limit float64) data.Float {
	if s.value != nil {
		if *s.value <= limit {
			s.context.AppendError(s.reference, service.ErrorValueNotGreaterThan(*s.value, limit))
		}
	}
	return s
}

func (s *StandardFloat) GreaterThanOrEqualTo(limit float64) data.Float {
	if s.value != nil {
		if *s.value < limit {
			s.context.AppendError(s.reference, service.ErrorValueNotGreaterThanOrEqualTo(*s.value, limit))
		}
	}
	return s
}

func (s *StandardFloat) InRange(lowerLimit float64, upperLimit float64) data.Float {
	if s.value != nil {
		if *s.value < lowerLimit || *s.value > upperLimit {
			s.context.AppendError(s.reference, service.ErrorValueNotInRange(*s.value, lowerLimit, upperLimit))
		}
	}
	return s
}

func (s *StandardFloat) OneOf(allowedValues []float64) data.Float {
	if s.value != nil {
		for _, possibleValue := range allowedValues {
			if possibleValue == *s.value {
				return s
			}
		}
		s.context.AppendError(s.reference, service.ErrorValueFloatNotOneOf(*s.value, allowedValues))
	}
	return s
}

func (s *StandardFloat) NotOneOf(disallowedValues []float64) data.Float {
	if s.value != nil {
		for _, possibleValue := range disallowedValues {
			if possibleValue == *s.value {
				s.context.AppendError(s.reference, service.ErrorValueFloatOneOf(*s.value, disallowedValues))
				return s
			}
		}
	}
	return s
}
