package pump

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	BolusAmountMaximumUnitsUnits        = "Units"
	BolusAmountMaximumValueUnitsMaximum = 100.0
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

func ParseBolusAmountMaximum(parser data.ObjectParser) *BolusAmountMaximum {
	if parser.Object() == nil {
		return nil
	}
	datum := NewBolusAmountMaximum()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewBolusAmountMaximum() *BolusAmountMaximum {
	return &BolusAmountMaximum{}
}

func (b *BolusAmountMaximum) Parse(parser data.ObjectParser) {
	b.Units = parser.ParseString("units")
	b.Value = parser.ParseFloat("value")
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
