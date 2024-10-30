package dexcom

import (
	"time"

	"github.com/tidepool-org/platform/structure"
)

type Moment struct {
	SystemTime  *Time `json:"systemTime,omitempty"`
	DisplayTime *Time `json:"displayTime,omitempty"`
}

func ParseMoment(parser structure.ObjectParser) *Moment {
	if !parser.Exists() {
		return nil
	}
	datum := NewMoment()
	parser.Parse(datum)
	return datum
}

func NewMoment() *Moment {
	return &Moment{}
}

func (m *Moment) Parse(parser structure.ObjectParser) {
	m.SystemTime = ParseTime(parser, "systemTime")
	m.DisplayTime = ParseTime(parser, "displayTime")
}

func (m *Moment) Validate(validator structure.Validator) {
	if systemTime := m.SystemTime.Raw(); systemTime != nil && !systemTime.Before(time.Now().Add(SystemTimeNowThreshold)) {
		validator.Logger().Warn("SystemTime is not before now with threshold")
	}

	validator.Time("systemTime", m.SystemTime.Raw()).Exists().NotZero()
	validator.Time("displayTime", m.DisplayTime.Raw()).Exists().NotZero()
}

func (m *Moment) SystemTimeRaw() *time.Time {
	if m.SystemTime == nil {
		return nil
	} else if systemTimeRaw := m.SystemTime.Raw(); systemTimeRaw == nil || systemTimeRaw.IsZero() {
		return nil
	} else {
		return systemTimeRaw
	}
}

type Moments []*Moment

func (m Moments) Compact() Moments {
	var moments Moments
	for _, moment := range m {
		if moment != nil {
			moments = append(moments, moment)
		}
	}
	return moments
}

type MomentsBySystemTimeRaw Moments

func (m MomentsBySystemTimeRaw) Len() int {
	return len(m)
}

// Sort with nils at end
func (m MomentsBySystemTimeRaw) Less(left int, right int) bool {
	if leftSystemTime := systemTimeRaw(m[left]); leftSystemTime == nil {
		return false
	} else if rightSystemTime := systemTimeRaw(m[right]); rightSystemTime == nil {
		return true
	} else {
		return leftSystemTime.Before(*rightSystemTime)
	}
}

func (m MomentsBySystemTimeRaw) Swap(left int, right int) {
	m[left], m[right] = m[right], m[left]
}

func systemTimeRaw(moment *Moment) *time.Time {
	if moment == nil {
		return nil
	} else {
		return moment.SystemTimeRaw()
	}
}
