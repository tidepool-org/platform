package food

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	EnergyKilojoulesPerKilocalorie = 4.1858
	EnergyUnitsCalories            = "calories"
	EnergyUnitsJoules              = "joules"
	EnergyUnitsKilocalories        = "kilocalories"
	EnergyUnitsKilojoules          = "kilojoules"
	EnergyValueCaloriesMaximum     = EnergyValueKilocaloriesMaximum * 1000.0
	EnergyValueCaloriesMinimum     = EnergyValueKilocaloriesMinimum * 1000.0
	EnergyValueJoulesMaximum       = EnergyValueKilojoulesMaximum * 1000.0
	EnergyValueJoulesMinimum       = EnergyValueKilojoulesMinimum * 1000.0
	EnergyValueKilocaloriesMaximum = 10000.0
	EnergyValueKilocaloriesMinimum = 0.0
	EnergyValueKilojoulesMaximum   = EnergyValueKilocaloriesMaximum * EnergyKilojoulesPerKilocalorie
	EnergyValueKilojoulesMinimum   = EnergyValueKilocaloriesMinimum * EnergyKilojoulesPerKilocalorie
)

func EnergyUnits() []string {
	return []string{
		EnergyUnitsCalories,
		EnergyUnitsJoules,
		EnergyUnitsKilocalories,
		EnergyUnitsKilojoules,
	}
}

type Energy struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseEnergy(parser data.ObjectParser) *Energy {
	if parser.Object() == nil {
		return nil
	}
	datum := NewEnergy()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewEnergy() *Energy {
	return &Energy{}
}

func (e *Energy) Parse(parser data.ObjectParser) {
	e.Units = parser.ParseString("units")
	e.Value = parser.ParseFloat("value")
}

func (e *Energy) Validate(validator structure.Validator) {
	validator.String("units", e.Units).Exists().OneOf(EnergyUnits()...)
	validator.Float64("value", e.Value).Exists().InRange(EnergyValueRangeForUnits(e.Units))
}

func (e *Energy) Normalize(normalizer data.Normalizer) {}

func EnergyValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case EnergyUnitsCalories:
			return EnergyValueCaloriesMinimum, EnergyValueCaloriesMaximum
		case EnergyUnitsJoules:
			return EnergyValueJoulesMinimum, EnergyValueJoulesMaximum
		case EnergyUnitsKilocalories:
			return EnergyValueKilocaloriesMinimum, EnergyValueKilocaloriesMaximum
		case EnergyUnitsKilojoules:
			return EnergyValueKilojoulesMinimum, EnergyValueKilojoulesMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
