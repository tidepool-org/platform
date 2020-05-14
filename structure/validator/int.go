package validator

import (
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Int struct {
	base  *structureBase.Base
	value *int
}

func NewInt(base *structureBase.Base, value *int) *Int {
	return &Int{
		base:  base,
		value: value,
	}
}

func (i *Int) Exists() structure.Int {
	if i.value == nil {
		i.base.ReportError(ErrorValueNotExists())
	}
	return i
}

func (i *Int) NotExists() structure.Int {
	if i.value != nil {
		i.base.ReportError(ErrorValueExists())
	}
	return i
}

func (i *Int) EqualTo(value int) structure.Int {
	if i.value != nil {
		if *i.value != value {
			i.base.ReportError(ErrorValueNotEqualTo(*i.value, value))
		}
	}
	return i
}

func (i *Int) NotEqualTo(value int) structure.Int {
	if i.value != nil {
		if *i.value == value {
			i.base.ReportError(ErrorValueEqualTo(*i.value, value))
		}
	}
	return i
}

func (i *Int) LessThan(limit int) structure.Int {
	if i.value != nil {
		if *i.value >= limit {
			i.base.ReportError(ErrorValueNotLessThan(*i.value, limit))
		}
	}
	return i
}

func (i *Int) LessThanOrEqualTo(limit int) structure.Int {
	if i.value != nil {
		if *i.value > limit {
			i.base.ReportError(ErrorValueNotLessThanOrEqualTo(*i.value, limit))
		}
	}
	return i
}

func (i *Int) GreaterThan(limit int) structure.Int {
	if i.value != nil {
		if *i.value <= limit {
			i.base.ReportError(ErrorValueNotGreaterThan(*i.value, limit))
		}
	}
	return i
}

func (i *Int) GreaterThanOrEqualTo(limit int) structure.Int {
	if i.value != nil {
		if *i.value < limit {
			i.base.ReportError(ErrorValueNotGreaterThanOrEqualTo(*i.value, limit))
		}
	}
	return i
}

func (i *Int) InRange(lowerLimit int, upperLimit int) structure.Int {
	if i.value != nil {
		if *i.value < lowerLimit || *i.value > upperLimit {
			i.base.ReportError(ErrorValueNotInRange(*i.value, lowerLimit, upperLimit))
		}
	}
	return i
}

func (i *Int) InRangeWarning(lowerLimit int, upperLimit int) structure.Int {
	if i.value != nil {
		if *i.value < lowerLimit || *i.value > upperLimit {
			i.base.ReportWarning(ErrorValueNotInRange(*i.value, lowerLimit, upperLimit))
		}
	}
	return i
}

func (i *Int) OneOf(allowedValues ...int) structure.Int {
	if i.value != nil {
		for _, allowedValue := range allowedValues {
			if allowedValue == *i.value {
				return i
			}
		}
		i.base.ReportError(ErrorValueIntNotOneOf(*i.value, allowedValues))
	}
	return i
}

func (i *Int) NotOneOf(disallowedValues ...int) structure.Int {
	if i.value != nil {
		for _, disallowedValue := range disallowedValues {
			if disallowedValue == *i.value {
				i.base.ReportError(ErrorValueIntOneOf(*i.value, disallowedValues))
				return i
			}
		}
	}
	return i
}

func (i *Int) Using(usingFunc structure.IntUsingFunc) structure.Int {
	if i.value != nil {
		if usingFunc != nil {
			usingFunc(*i.value, i.base)
		}
	}
	return i
}
