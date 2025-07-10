package validator

import (
	"time"

	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Duration struct {
	base  *structureBase.Base
	value *time.Duration
}

func NewDuration(base *structureBase.Base, value *time.Duration) *Duration {
	return &Duration{
		base:  base,
		value: value,
	}
}

func (d *Duration) Exists() structure.Duration {
	if d.value == nil {
		d.base.ReportError(ErrorValueNotExists())
	}
	return d
}

func (d *Duration) NotExists() structure.Duration {
	if d.value != nil {
		d.base.ReportError(ErrorValueExists())
	}
	return d
}

func (d *Duration) EqualTo(value time.Duration) structure.Duration {
	if d.value != nil {
		if *d.value != value {
			d.base.ReportError(ErrorValueNotEqualTo(*d.value, value))
		}
	}
	return d
}

func (d *Duration) NotEqualTo(value time.Duration) structure.Duration {
	if d.value != nil {
		if *d.value == value {
			d.base.ReportError(ErrorValueEqualTo(*d.value, value))
		}
	}
	return d
}

func (d *Duration) LessThan(limit time.Duration) structure.Duration {
	if d.value != nil {
		if *d.value >= limit {
			d.base.ReportError(ErrorValueNotLessThan(*d.value, limit))
		}
	}
	return d
}

func (d *Duration) LessThanOrEqualTo(limit time.Duration) structure.Duration {
	if d.value != nil {
		if *d.value > limit {
			d.base.ReportError(ErrorValueNotLessThanOrEqualTo(*d.value, limit))
		}
	}
	return d
}

func (d *Duration) GreaterThan(limit time.Duration) structure.Duration {
	if d.value != nil {
		if *d.value <= limit {
			d.base.ReportError(ErrorValueNotGreaterThan(*d.value, limit))
		}
	}
	return d
}

func (d *Duration) GreaterThanOrEqualTo(limit time.Duration) structure.Duration {
	if d.value != nil {
		if *d.value < limit {
			d.base.ReportError(ErrorValueNotGreaterThanOrEqualTo(*d.value, limit))
		}
	}
	return d
}

func (d *Duration) InRange(lowerLimit time.Duration, upperLimit time.Duration) structure.Duration {
	if d.value != nil {
		if !structure.InRange(*d.value, lowerLimit, upperLimit) {
			d.base.ReportError(ErrorValueNotInRange(*d.value, lowerLimit, upperLimit))
		}
	}
	return d
}

func (d *Duration) OneOf(allowedValues ...time.Duration) structure.Duration {
	if d.value != nil {
		for _, allowedValue := range allowedValues {
			if allowedValue == *d.value {
				return d
			}
		}
		d.base.ReportError(ErrorValueDurationNotOneOf(*d.value, allowedValues))
	}
	return d
}

func (d *Duration) NotOneOf(disallowedValues ...time.Duration) structure.Duration {
	if d.value != nil {
		for _, disallowedValue := range disallowedValues {
			if disallowedValue == *d.value {
				d.base.ReportError(ErrorValueDurationOneOf(*d.value, disallowedValues))
				return d
			}
		}
	}
	return d
}

func (f *Duration) Using(usingFunc structure.DurationUsingFunc) structure.Duration {
	if f.value != nil {
		if usingFunc != nil {
			usingFunc(*f.value, f.base)
		}
	}
	return f
}
