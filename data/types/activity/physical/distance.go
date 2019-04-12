package physical

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	DistanceFeetPerMile            = 5280.0
	DistanceKilometersPerMile      = 1.609344
	DistanceMetersPerMile          = 1609.344
	DistanceUnitsFeet              = "feet"
	DistanceUnitsKilometers        = "kilometers"
	DistanceUnitsMeters            = "meters"
	DistanceUnitsMiles             = "miles"
	DistanceUnitsYards             = "yards"
	DistanceValueFeetMaximum       = DistanceValueMilesMaximum * DistanceFeetPerMile
	DistanceValueFeetMinimum       = DistanceValueMilesMinimum * DistanceFeetPerMile
	DistanceValueKilometersMaximum = DistanceValueMilesMaximum * DistanceKilometersPerMile
	DistanceValueKilometersMinimum = DistanceValueMilesMinimum * DistanceKilometersPerMile
	DistanceValueMetersMaximum     = DistanceValueMilesMaximum * DistanceMetersPerMile
	DistanceValueMetersMinimum     = DistanceValueMilesMinimum * DistanceMetersPerMile
	DistanceValueMilesMaximum      = 100.0
	DistanceValueMilesMinimum      = 0.0
	DistanceValueYardsMaximum      = DistanceValueMilesMaximum * DistanceYardsPerMile
	DistanceValueYardsMinimum      = DistanceValueMilesMinimum * DistanceYardsPerMile
	DistanceYardsPerMile           = 1760.0
)

func DistanceUnits() []string {
	return []string{
		DistanceUnitsFeet,
		DistanceUnitsKilometers,
		DistanceUnitsMeters,
		DistanceUnitsMiles,
		DistanceUnitsYards,
	}
}

type Distance struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseDistance(parser structure.ObjectParser) *Distance {
	if !parser.Exists() {
		return nil
	}
	datum := NewDistance()
	parser.Parse(datum)
	return datum
}

func NewDistance() *Distance {
	return &Distance{}
}

func (d *Distance) Parse(parser structure.ObjectParser) {
	d.Units = parser.String("units")
	d.Value = parser.Float64("value")
}

func (d *Distance) Validate(validator structure.Validator) {
	validator.String("units", d.Units).Exists().OneOf(DistanceUnits()...)
	validator.Float64("value", d.Value).Exists().InRange(DistanceValueRangeForUnits(d.Units))
}

func (d *Distance) Normalize(normalizer data.Normalizer) {}

func DistanceValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case DistanceUnitsFeet:
			return DistanceValueFeetMinimum, DistanceValueFeetMaximum
		case DistanceUnitsKilometers:
			return DistanceValueKilometersMinimum, DistanceValueKilometersMaximum
		case DistanceUnitsMeters:
			return DistanceValueMetersMinimum, DistanceValueMetersMaximum
		case DistanceUnitsMiles:
			return DistanceValueMilesMinimum, DistanceValueMilesMaximum
		case DistanceUnitsYards:
			return DistanceValueYardsMinimum, DistanceValueYardsMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
