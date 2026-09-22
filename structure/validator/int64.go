package validator

import (
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Int64 struct {
	base  *structureBase.Base
	value *int64
}

func NewInt64(base *structureBase.Base, value *int64) *Int64 {
	return &Int64{
		base:  base,
		value: value,
	}
}

func (i *Int64) Exists() structure.Int64 {
	if i.value == nil {
		i.base.ReportError(ErrorValueNotExists())
	}
	return i
}

func (i *Int64) NotExists() structure.Int64 {
	if i.value != nil {
		i.base.ReportError(ErrorValueExists())
	}
	return i
}

func (i *Int64) EqualTo(value int64) structure.Int64 {
	if i.value != nil {
		if *i.value != value {
			i.base.ReportError(ErrorValueNotEqualTo(*i.value, value))
		}
	}
	return i
}

func (i *Int64) NotEqualTo(value int64) structure.Int64 {
	if i.value != nil {
		if *i.value == value {
			i.base.ReportError(ErrorValueEqualTo(*i.value, value))
		}
	}
	return i
}

func (i *Int64) LessThan(limit int64) structure.Int64 {
	if i.value != nil {
		if *i.value >= limit {
			i.base.ReportError(ErrorValueNotLessThan(*i.value, limit))
		}
	}
	return i
}

func (i *Int64) LessThanOrEqualTo(limit int64) structure.Int64 {
	if i.value != nil {
		if *i.value > limit {
			i.base.ReportError(ErrorValueNotLessThanOrEqualTo(*i.value, limit))
		}
	}
	return i
}

func (i *Int64) GreaterThan(limit int64) structure.Int64 {
	if i.value != nil {
		if *i.value <= limit {
			i.base.ReportError(ErrorValueNotGreaterThan(*i.value, limit))
		}
	}
	return i
}

func (i *Int64) GreaterThanOrEqualTo(limit int64) structure.Int64 {
	if i.value != nil {
		if *i.value < limit {
			i.base.ReportError(ErrorValueNotGreaterThanOrEqualTo(*i.value, limit))
		}
	}
	return i
}

func (i *Int64) InRange(lowerLimit int64, upperLimit int64) structure.Int64 {
	if i.value != nil {
		if !structure.InRange(*i.value, lowerLimit, upperLimit) {
			i.base.ReportError(ErrorValueNotInRange(*i.value, lowerLimit, upperLimit))
		}
	}
	return i
}

func (i *Int64) OneOf(allowedValues ...int64) structure.Int64 {
	if i.value != nil {
		for _, allowedValue := range allowedValues {
			if allowedValue == *i.value {
				return i
			}
		}
		i.base.ReportError(ErrorValueInt64NotOneOf(*i.value, allowedValues))
	}
	return i
}

func (i *Int64) NotOneOf(disallowedValues ...int64) structure.Int64 {
	if i.value != nil {
		for _, disallowedValue := range disallowedValues {
			if disallowedValue == *i.value {
				i.base.ReportError(ErrorValueInt64OneOf(*i.value, disallowedValues))
				return i
			}
		}
	}
	return i
}

func (i *Int64) Using(usingFunc structure.Int64UsingFunc) structure.Int64 {
	if i.value != nil {
		if usingFunc != nil {
			usingFunc(*i.value, i.base)
		}
	}
	return i
}
