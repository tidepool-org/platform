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
	validator.Time("systemTime", m.SystemTime.Raw()).Exists().NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", m.DisplayTime.Raw()).Exists().NotZero()
}

func (m *Moment) Normalize(normalizer structure.Normalizer) {}

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

func (m Moments) Compact() []*Moment {
	var moments Moments
	for _, moment := range m {
		if moment != nil {
			moments = append(moments, moment)
		}
	}
	return moments
}

type BySystemTimeRaw []*Moment

func (b BySystemTimeRaw) Len() int {
	return len(b)
}

// Sort with nils at end
func (b BySystemTimeRaw) Less(left int, right int) bool {
	if leftSystemTime := systemTimeRaw(b[left]); leftSystemTime == nil {
		return false
	} else if rightSystemTime := systemTimeRaw(b[right]); rightSystemTime == nil {
		return true
	} else {
		return leftSystemTime.Before(*rightSystemTime)
	}
}

func (b BySystemTimeRaw) Swap(left int, right int) {
	b[left], b[right] = b[right], b[left]
}

func systemTimeRaw(moment *Moment) *time.Time {
	if moment == nil {
		return nil
	} else {
		return moment.SystemTimeRaw()
	}
}
