package location

import (
	"math"

	"github.com/tidepool-org/platform/structure"
)

const (
	ElevationUnitsFeet          = "feet"
	ElevationUnitsMeters        = "meters"
	ElevationValueFeetMaximum   = ElevationValueMetersMaximum / 0.3048
	ElevationValueFeetMinimum   = ElevationValueMetersMinimum / 0.3048
	ElevationValueMetersMaximum = 1000000.0
	ElevationValueMetersMinimum = -20000.0
)

func ElevationUnits() []string {
	return []string{
		ElevationUnitsFeet,
		ElevationUnitsMeters,
	}
}

type Elevation struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseElevation(parser structure.ObjectParser) *Elevation {
	if !parser.Exists() {
		return nil
	}
	datum := NewElevation()
	parser.Parse(datum)
	return datum
}

func NewElevation() *Elevation {
	return &Elevation{}
}

func (e *Elevation) Parse(parser structure.ObjectParser) {
	e.Units = parser.String("units")
	e.Value = parser.Float64("value")
}

func (e *Elevation) Validate(validator structure.Validator) {
	validator.String("units", e.Units).Exists().OneOf(ElevationUnits()...)
	validator.Float64("value", e.Value).Exists().InRange(ElevationValueRangeForUnits(e.Units))
}

func ElevationValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case ElevationUnitsFeet:
			return ElevationValueFeetMinimum, ElevationValueFeetMaximum
		case ElevationUnitsMeters:
			return ElevationValueMetersMinimum, ElevationValueMetersMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
