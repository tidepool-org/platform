package pump

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	BolusCalculatorInsulinDurationHoursMaximum   = 10.0
	BolusCalculatorInsulinDurationHoursMinimum   = 0.0
	BolusCalculatorInsulinDurationMinutesMaximum = BolusCalculatorInsulinDurationHoursMaximum * 60.0
	BolusCalculatorInsulinDurationMinutesMinimum = BolusCalculatorInsulinDurationHoursMinimum * 60.0
	BolusCalculatorInsulinDurationSecondsMaximum = BolusCalculatorInsulinDurationMinutesMaximum * 60.0
	BolusCalculatorInsulinDurationSecondsMinimum = BolusCalculatorInsulinDurationMinutesMinimum * 60.0
	BolusCalculatorInsulinUnitsHours             = "hours"
	BolusCalculatorInsulinUnitsMinutes           = "minutes"
	BolusCalculatorInsulinUnitsSeconds           = "seconds"
)

func BolusCalculatorInsulinUnits() []string {
	return []string{
		BolusCalculatorInsulinUnitsHours,
		BolusCalculatorInsulinUnitsMinutes,
		BolusCalculatorInsulinUnitsSeconds,
	}
}

type BolusCalculatorInsulin struct {
	Duration *float64 `json:"duration,omitempty" bson:"duration,omitempty"`
	Units    *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseBolusCalculatorInsulin(parser structure.ObjectParser) *BolusCalculatorInsulin {
	if !parser.Exists() {
		return nil
	}
	datum := NewBolusCalculatorInsulin()
	parser.Parse(datum)
	return datum
}

func NewBolusCalculatorInsulin() *BolusCalculatorInsulin {
	return &BolusCalculatorInsulin{}
}

func (b *BolusCalculatorInsulin) Parse(parser structure.ObjectParser) {
	b.Duration = parser.Float64("duration")
	b.Units = parser.String("units")
}

func (b *BolusCalculatorInsulin) Validate(validator structure.Validator) {
	validator.Float64("duration", b.Duration).Exists().InRange(BolusCalculatorInsulinDurationRangeForUnits(b.Units))
	validator.String("units", b.Units).Exists().OneOf(BolusCalculatorInsulinUnits()...)
}

func (b *BolusCalculatorInsulin) Normalize(normalizer data.Normalizer) {}

func BolusCalculatorInsulinDurationRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case BolusCalculatorInsulinUnitsHours:
			return BolusCalculatorInsulinDurationHoursMinimum, BolusCalculatorInsulinDurationHoursMaximum
		case BolusCalculatorInsulinUnitsMinutes:
			return BolusCalculatorInsulinDurationMinutesMinimum, BolusCalculatorInsulinDurationMinutesMaximum
		case BolusCalculatorInsulinUnitsSeconds:
			return BolusCalculatorInsulinDurationSecondsMinimum, BolusCalculatorInsulinDurationSecondsMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
