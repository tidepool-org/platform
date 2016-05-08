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

type StandardFloat struct {
	context   data.Context
	reference interface{}
	value     *float64
}

func NewStandardFloat(context data.Context, reference interface{}, value *float64) *StandardFloat {
	return &StandardFloat{
		context:   context,
		reference: reference,
		value:     value,
	}
}

func (s *StandardFloat) Exists() data.Float {
	if s.value == nil {
		s.context.AppendError(s.reference, ErrorValueDoesNotExist())
	}
	return s
}

func (s *StandardFloat) EqualTo(limit float64) data.Float {
	if s.value != nil {
		if *s.value != limit {
			s.context.AppendError(s.reference, ErrorValueNotEqualTo(*s.value, limit))
		}
	}
	return s
}

func (s *StandardFloat) NotEqualTo(limit float64) data.Float {
	if s.value != nil {
		if *s.value == limit {
			s.context.AppendError(s.reference, ErrorValueEqualTo(*s.value, limit))
		}
	}
	return s
}

func (s *StandardFloat) LessThan(limit float64) data.Float {
	if s.value != nil {
		if *s.value >= limit {
			s.context.AppendError(s.reference, ErrorValueNotLessThan(*s.value, limit))
		}
	}
	return s
}

func (s *StandardFloat) LessThanOrEqualTo(limit float64) data.Float {
	if s.value != nil {
		if *s.value > limit {
			s.context.AppendError(s.reference, ErrorValueNotLessThanOrEqual(*s.value, limit))
		}
	}
	return s
}

func (s *StandardFloat) GreaterThan(limit float64) data.Float {
	if s.value != nil {
		if *s.value <= limit {
			s.context.AppendError(s.reference, ErrorValueNotGreaterThan(*s.value, limit))
		}
	}
	return s
}

func (s *StandardFloat) GreaterThanOrEqualTo(limit float64) data.Float {
	if s.value != nil {
		if *s.value < limit {
			s.context.AppendError(s.reference, ErrorValueNotGreaterThanOrEqual(*s.value, limit))
		}
	}
	return s
}

func (s *StandardFloat) InRange(lowerlimit float64, upperLimit float64) data.Float {
	if s.value != nil {
		if *s.value < lowerlimit || *s.value > upperLimit {
			s.context.AppendError(s.reference, ErrorFloatNotInRange(*s.value, lowerlimit, upperLimit))
		}
	}
	return s
}

func (s *StandardFloat) OneOf(possibleValues []float64) data.Float {
	if s.value != nil {
		for _, possibleValue := range possibleValues {
			if possibleValue == *s.value {
				return s
			}
		}
		s.context.AppendError(s.reference, ErrorFloatNotOneOf(*s.value, possibleValues))
	}
	return s
}
