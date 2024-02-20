package dexcom

import (
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
	m.SystemTime = TimeFromString(parser.String("systemTime"))
	m.DisplayTime = TimeFromString(parser.String("displayTime"))
}

func (m *Moment) Validate(validator structure.Validator) {
	validator.Time("systemTime", m.SystemTime.Raw()).Exists().NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", m.DisplayTime.Raw()).Exists().NotZero()
}

func (m *Moment) Normalize(normalizer structure.Normalizer) {}
