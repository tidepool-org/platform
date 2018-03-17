package validator

import (
	"time"

	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type Time struct {
	base  *structureBase.Base
	value *time.Time
}

func NewTime(base *structureBase.Base, value *time.Time) *Time {
	return &Time{
		base:  base,
		value: value,
	}
}

func (t *Time) Exists() structure.Time {
	if t.value == nil {
		t.base.ReportError(ErrorValueNotExists())
	}
	return t
}

func (t *Time) NotExists() structure.Time {
	if t.value != nil {
		t.base.ReportError(ErrorValueExists())
	}
	return t
}

func (t *Time) Zero() structure.Time {
	if t.value != nil {
		if !(*t.value).IsZero() {
			t.base.ReportError(ErrorValueNotEmpty())
		}
	}
	return t
}

func (t *Time) NotZero() structure.Time {
	if t.value != nil {
		if (*t.value).IsZero() {
			t.base.ReportError(ErrorValueEmpty())
		}
	}
	return t
}

func (t *Time) After(limit time.Time) structure.Time {
	if t.value != nil {
		if (*t.value).Before(limit) {
			t.base.ReportError(ErrorValueTimeNotAfter(*t.value, limit))
		}
	}
	return t
}

func (t *Time) AfterNow(threshold time.Duration) structure.Time {
	if t.value != nil {
		if (*t.value).Before(time.Now().Add(-threshold)) {
			t.base.ReportError(ErrorValueTimeNotAfterNow(*t.value))
		}
	}
	return t
}

func (t *Time) Before(limit time.Time) structure.Time {
	if t.value != nil {
		if (*t.value).After(limit) {
			t.base.ReportError(ErrorValueTimeNotBefore(*t.value, limit))
		}
	}
	return t
}

func (t *Time) BeforeNow(threshold time.Duration) structure.Time {
	if t.value != nil {
		if (*t.value).After(time.Now().Add(threshold)) {
			t.base.ReportError(ErrorValueTimeNotBeforeNow(*t.value))
		}
	}
	return t
}

func (t *Time) Using(using func(value time.Time, errorReporter structure.ErrorReporter)) structure.Time {
	if t.value != nil {
		if using != nil {
			using(*t.value, t.base)
		}
	}
	return t
}
