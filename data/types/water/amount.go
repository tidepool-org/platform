package water

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	AmountLitersPerGallon         = 3.7854118
	AmountOuncesPerGallon         = 128.0
	AmountUnitsGallons            = "gallons"
	AmountUnitsLiters             = "liters"
	AmountUnitsMilliliters        = "milliliters"
	AmountUnitsOunces             = "ounces"
	AmountValueGallonsMaximum     = 10.0
	AmountValueGallonsMinimum     = 0.0
	AmountValueLitersMaximum      = AmountValueGallonsMaximum * AmountLitersPerGallon
	AmountValueLitersMinimum      = AmountValueGallonsMinimum * AmountLitersPerGallon
	AmountValueMillilitersMaximum = AmountValueLitersMaximum * 1000.0
	AmountValueMillilitersMinimum = AmountValueLitersMinimum * 1000.0
	AmountValueOuncesMaximum      = AmountValueGallonsMaximum * AmountOuncesPerGallon
	AmountValueOuncesMinimum      = AmountValueGallonsMinimum * AmountOuncesPerGallon
)

func AmountUnits() []string {
	return []string{
		AmountUnitsGallons,
		AmountUnitsLiters,
		AmountUnitsMilliliters,
		AmountUnitsOunces,
	}
}

type Amount struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseAmount(parser structure.ObjectParser) *Amount {
	if !parser.Exists() {
		return nil
	}
	datum := NewAmount()
	parser.Parse(datum)
	return datum
}

func NewAmount() *Amount {
	return &Amount{}
}

func (a *Amount) Parse(parser structure.ObjectParser) {
	a.Units = parser.String("units")
	a.Value = parser.Float64("value")
}

func (a *Amount) Validate(validator structure.Validator) {
	validator.String("units", a.Units).Exists().OneOf(AmountUnits()...)
	validator.Float64("value", a.Value).Exists().InRange(AmountValueRangeForUnits(a.Units))
}

func (a *Amount) Normalize(normalizer data.Normalizer) {}

func AmountValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case AmountUnitsGallons:
			return AmountValueGallonsMinimum, AmountValueGallonsMaximum
		case AmountUnitsLiters:
			return AmountValueLitersMinimum, AmountValueLitersMaximum
		case AmountUnitsMilliliters:
			return AmountValueMillilitersMinimum, AmountValueMillilitersMaximum
		case AmountUnitsOunces:
			return AmountValueOuncesMinimum, AmountValueOuncesMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
