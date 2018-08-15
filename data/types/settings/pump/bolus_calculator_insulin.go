package pump

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	BolusCalculatorInsulinDurationHoursMaximum = 10.0
	BolusCalculatorInsulinDurationHoursMinimum = 0.0
	BolusCalculatorInsulinUnitsHours           = "hours"
)

func BolusCalculatorInsulinUnits() []string {
	return []string{
		BolusCalculatorInsulinUnitsHours,
	}
}

type BolusCalculatorInsulin struct {
	Duration *float64 `json:"duration,omitempty" bson:"duration,omitempty"`
	Units    *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseBolusCalculatorInsulin(parser data.ObjectParser) *BolusCalculatorInsulin {
	if parser.Object() == nil {
		return nil
	}
	datum := NewBolusCalculatorInsulin()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewBolusCalculatorInsulin() *BolusCalculatorInsulin {
	return &BolusCalculatorInsulin{}
}

func (b *BolusCalculatorInsulin) Parse(parser data.ObjectParser) {
	b.Duration = parser.ParseFloat("duration")
	b.Units = parser.ParseString("units")
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
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
