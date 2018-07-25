package pump

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	BasalRateMaximumUnitsUnitsPerHour        = "Units/hour"
	BasalRateMaximumValueUnitsPerHourMaximum = 100.0
	BasalRateMaximumValueUnitsPerHourMinimum = 0.0
)

func BasalRateMaximumUnits() []string {
	return []string{
		BasalRateMaximumUnitsUnitsPerHour,
	}
}

type BasalRateMaximum struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseBasalRateMaximum(parser data.ObjectParser) *BasalRateMaximum {
	if parser.Object() == nil {
		return nil
	}
	datum := NewBasalRateMaximum()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewBasalRateMaximum() *BasalRateMaximum {
	return &BasalRateMaximum{}
}

func (b *BasalRateMaximum) Parse(parser data.ObjectParser) {
	b.Units = parser.ParseString("units")
	b.Value = parser.ParseFloat("value")
}

func (b *BasalRateMaximum) Validate(validator structure.Validator) {
	validator.String("units", b.Units).Exists().OneOf(BasalRateMaximumUnits()...)
	validator.Float64("value", b.Value).Exists().InRange(BasalRateMaximumValueRangeForUnits(b.Units))
}

func (b *BasalRateMaximum) Normalize(normalizer data.Normalizer) {}

func BasalRateMaximumValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case BasalRateMaximumUnitsUnitsPerHour:
			return BasalRateMaximumValueUnitsPerHourMinimum, BasalRateMaximumValueUnitsPerHourMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
