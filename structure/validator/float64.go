package validator

import (
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Float64 struct {
	base  *structureBase.Base
	value *float64
}

func NewFloat64(base *structureBase.Base, value *float64) *Float64 {
	return &Float64{
		base:  base,
		value: value,
	}
}

func (f *Float64) Exists() structure.Float64 {
	if f.value == nil {
		f.base.ReportError(ErrorValueNotExists())
	}
	return f
}

func (f *Float64) NotExists() structure.Float64 {
	if f.value != nil {
		f.base.ReportError(ErrorValueExists())
	}
	return f
}

func (f *Float64) EqualTo(value float64) structure.Float64 {
	if f.value != nil {
		if *f.value != value {
			f.base.ReportError(ErrorValueNotEqualTo(*f.value, value))
		}
	}
	return f
}

func (f *Float64) NotEqualTo(value float64) structure.Float64 {
	if f.value != nil {
		if *f.value == value {
			f.base.ReportError(ErrorValueEqualTo(*f.value, value))
		}
	}
	return f
}

func (f *Float64) LessThan(limit float64) structure.Float64 {
	if f.value != nil {
		if *f.value >= limit {
			f.base.ReportError(ErrorValueNotLessThan(*f.value, limit))
		}
	}
	return f
}

func (f *Float64) LessThanOrEqualTo(limit float64) structure.Float64 {
	if f.value != nil {
		if *f.value > limit {
			f.base.ReportError(ErrorValueNotLessThanOrEqualTo(*f.value, limit))
		}
	}
	return f
}

func (f *Float64) GreaterThan(limit float64) structure.Float64 {
	if f.value != nil {
		if *f.value <= limit {
			f.base.ReportError(ErrorValueNotGreaterThan(*f.value, limit))
		}
	}
	return f
}

func (f *Float64) GreaterThanOrEqualTo(limit float64) structure.Float64 {
	if f.value != nil {
		if *f.value < limit {
			f.base.ReportError(ErrorValueNotGreaterThanOrEqualTo(*f.value, limit))
		}
	}
	return f
}

func (f *Float64) InRange(lowerLimit float64, upperLimit float64) structure.Float64 {
	if f.value != nil {
		if *f.value < lowerLimit || *f.value > upperLimit {
			f.base.ReportError(ErrorValueNotInRange(*f.value, lowerLimit, upperLimit))
		}
	}
	return f
}

func (f *Float64) OneOf(allowedValues ...float64) structure.Float64 {
	if f.value != nil {
		for _, allowedValue := range allowedValues {
			if allowedValue == *f.value {
				return f
			}
		}
		f.base.ReportError(ErrorValueFloat64NotOneOf(*f.value, allowedValues))
	}
	return f
}

func (f *Float64) NotOneOf(disallowedValues ...float64) structure.Float64 {
	if f.value != nil {
		for _, disallowedValue := range disallowedValues {
			if disallowedValue == *f.value {
				f.base.ReportError(ErrorValueFloat64OneOf(*f.value, disallowedValues))
				return f
			}
		}
	}
	return f
}
