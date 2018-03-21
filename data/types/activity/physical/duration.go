package physical

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	DurationUnitsHours          = "hours"
	DurationUnitsMinutes        = "minutes"
	DurationUnitsSeconds        = "seconds"
	DurationValueHoursMaximum   = 168.0
	DurationValueHoursMinimum   = 0.0
	DurationValueMinutesMaximum = DurationValueHoursMaximum * 60.0
	DurationValueMinutesMinimum = DurationValueHoursMinimum * 60.0
	DurationValueSecondsMaximum = DurationValueMinutesMaximum * 60.0
	DurationValueSecondsMinimum = DurationValueMinutesMinimum * 60.0
)

func DurationUnits() []string {
	return []string{
		DurationUnitsHours,
		DurationUnitsMinutes,
		DurationUnitsSeconds,
	}
}

type Duration struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseDuration(parser data.ObjectParser) *Duration {
	if parser.Object() == nil {
		return nil
	}
	datum := NewDuration()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewDuration() *Duration {
	return &Duration{}
}

func (d *Duration) Parse(parser data.ObjectParser) {
	d.Units = parser.ParseString("units")
	d.Value = parser.ParseFloat("value")
}

func (d *Duration) Validate(validator structure.Validator) {
	validator.String("units", d.Units).Exists().OneOf(DurationUnits()...)
	validator.Float64("value", d.Value).Exists().InRange(DurationValueRangeForUnits(d.Units))
}

func (d *Duration) Normalize(normalizer data.Normalizer) {}

func DurationValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case DurationUnitsHours:
			return DurationValueHoursMinimum, DurationValueHoursMaximum
		case DurationUnitsMinutes:
			return DurationValueMinutesMinimum, DurationValueMinutesMaximum
		case DurationUnitsSeconds:
			return DurationValueSecondsMinimum, DurationValueSecondsMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
