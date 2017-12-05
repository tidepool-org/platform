package physical

import (
	"github.com/tidepool-org/platform/data"
)

const (
	UnitsHours   = "hours"
	UnitsMinutes = "minutes"
	UnitsSeconds = "seconds"

	ValueDaysMaximum    = 7
	ValueHoursMaximum   = ValueDaysMaximum * MinutesPerHour
	ValueMinutesMaximum = ValueHoursMaximum * MinutesPerHour
	ValueSecondsMaximum = ValueMinutesMaximum * SecondsPerMinute

	HoursPerDay      = 24
	MinutesPerHour   = 60
	SecondsPerMinute = 60
)

type Duration struct {
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func NewDuration() *Duration {
	return &Duration{}
}

func (d *Duration) Parse(parser data.ObjectParser) {
	d.Value = parser.ParseFloat("value")
	d.Units = parser.ParseString("units")
}

func (d *Duration) Validate(validator data.Validator) {
	valueValidator := validator.ValidateFloat("value", d.Value)
	valueValidator.Exists()
	if d.Units != nil {
		switch *d.Units {
		case UnitsHours:
			valueValidator.InRange(0, ValueHoursMaximum)
		case UnitsMinutes:
			valueValidator.InRange(0, ValueMinutesMaximum)
		case UnitsSeconds:
			valueValidator.InRange(0, ValueSecondsMaximum)
		}
	}
	validator.ValidateString("units", d.Units).Exists().OneOf([]string{UnitsHours, UnitsMinutes, UnitsSeconds})
}

func (d *Duration) Normalize(normalizer data.Normalizer) {}

func ParseDuration(parser data.ObjectParser) *Duration {
	if parser.Object() == nil {
		return nil
	}

	duration := NewDuration()
	duration.Parse(parser)
	parser.ProcessNotParsed()

	return duration
}
