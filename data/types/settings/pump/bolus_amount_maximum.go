package pump

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	BolusAmountMaximumUnitsUnits        = "Units"
	BolusAmountMaximumValueUnitsMaximum = 250.0
	BolusAmountMaximumValueUnitsMinimum = 0.0
)

func BolusAmountMaximumUnits() []string {
	return []string{
		BolusAmountMaximumUnitsUnits,
	}
}

type BolusAmountMaximum struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseBolusAmountMaximum(parser structure.ObjectParser) *BolusAmountMaximum {
	if !parser.Exists() {
		return nil
	}
	datum := NewBolusAmountMaximum()
	parser.Parse(datum)
	return datum
}

func NewBolusAmountMaximum() *BolusAmountMaximum {
	return &BolusAmountMaximum{}
}

func (b *BolusAmountMaximum) Parse(parser structure.ObjectParser) {
	b.Units = parser.String("units")
	b.Value = parser.Float64("value")
}

func (b *BolusAmountMaximum) Validate(validator structure.Validator) {
	validator.String("units", b.Units).Exists().OneOf(BolusAmountMaximumUnits()...)
	validator.Float64("value", b.Value).Exists().InRange(BolusAmountMaximumValueRangeForUnits(b.Units))
}

func (b *BolusAmountMaximum) Normalize(normalizer data.Normalizer) {}

func BolusAmountMaximumValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case BolusAmountMaximumUnitsUnits:
			return BolusAmountMaximumValueUnitsMinimum, BolusAmountMaximumValueUnitsMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
