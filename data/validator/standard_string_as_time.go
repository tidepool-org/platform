package validator

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/service"
)

type StandardStringAsTime struct {
	context     data.Context
	reference   interface{}
	stringValue *string
	timeValue   *time.Time
	timeLayout  string
}

func NewStandardStringAsTime(context data.Context, reference interface{}, stringValue *string, timeLayout string) *StandardStringAsTime {
	if context == nil {
		return nil
	}
	if timeLayout == "" {
		return nil
	}

	standardStringAsTime := &StandardStringAsTime{
		context:     context,
		reference:   reference,
		stringValue: stringValue,
		timeLayout:  timeLayout,
	}
	standardStringAsTime.parse()
	return standardStringAsTime
}

func (s *StandardStringAsTime) Exists() data.Time {
	if s.stringValue == nil {
		s.context.AppendError(s.reference, service.ErrorValueNotExists())
	}
	return s
}

func (s *StandardStringAsTime) NotExists() data.Time {
	if s.stringValue != nil {
		s.context.AppendError(s.reference, service.ErrorValueExists())
	}
	return s
}

func (s *StandardStringAsTime) After(limit time.Time) data.Time {
	if s.timeValue != nil {
		if !s.timeValue.After(limit) {
			s.context.AppendError(s.reference, service.ErrorValueTimeNotAfter(*s.timeValue, limit, s.timeLayout))
		}
	}
	return s
}

func (s *StandardStringAsTime) AfterNow(threshold time.Duration) data.Time {
	if s.timeValue != nil {
		if !s.timeValue.After(time.Now().Add(threshold)) {
			s.context.AppendError(s.reference, service.ErrorValueTimeNotAfterNow(*s.timeValue, s.timeLayout))
		}
	}
	return s
}

func (s *StandardStringAsTime) Before(limit time.Time) data.Time {
	if s.timeValue != nil {
		if !s.timeValue.Before(limit) {
			s.context.AppendError(s.reference, service.ErrorValueTimeNotBefore(*s.timeValue, limit, s.timeLayout))
		}
	}
	return s
}

func (s *StandardStringAsTime) BeforeNow(threshold time.Duration) data.Time {
	if s.timeValue != nil {
		if !s.timeValue.Before(time.Now().Add(-threshold)) {
			s.context.AppendError(s.reference, service.ErrorValueTimeNotBeforeNow(*s.timeValue, s.timeLayout))
		}
	}
	return s
}

func (s *StandardStringAsTime) parse() {
	if s.stringValue != nil {
		if timeValue, err := time.Parse(s.timeLayout, *s.stringValue); err != nil {
			s.context.AppendError(s.reference, service.ErrorValueTimeNotValid(*s.stringValue, s.timeLayout))
		} else {
			s.timeValue = &timeValue
		}
	}
}
