package physical

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	ElevationChangeMetersPerFoot      = 0.3048
	ElevationChangeUnitsFeet          = "feet"
	ElevationChangeUnitsMeters        = "meters"
	ElevationChangeValueFeetMaximum   = 52800.0
	ElevationChangeValueFeetMinimum   = 0.0
	ElevationChangeValueMetersMaximum = ElevationChangeValueFeetMaximum * ElevationChangeMetersPerFoot
	ElevationChangeValueMetersMinimum = ElevationChangeValueFeetMinimum * ElevationChangeMetersPerFoot
)

func ElevationChangeUnits() []string {
	return []string{
		ElevationChangeUnitsFeet,
		ElevationChangeUnitsMeters,
	}
}

type ElevationChange struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseElevationChange(parser structure.ObjectParser) *ElevationChange {
	if !parser.Exists() {
		return nil
	}
	datum := NewElevationChange()
	parser.Parse(datum)
	return datum
}

func NewElevationChange() *ElevationChange {
	return &ElevationChange{}
}

func (e *ElevationChange) Parse(parser structure.ObjectParser) {
	e.Units = parser.String("units")
	e.Value = parser.Float64("value")
}

func (e *ElevationChange) Validate(validator structure.Validator) {
	validator.String("units", e.Units).Exists().OneOf(ElevationChangeUnits()...)
	validator.Float64("value", e.Value).Exists().InRange(ElevationChangeValueRangeForUnits(e.Units))
}

func (e *ElevationChange) Normalize(normalizer data.Normalizer) {}

func ElevationChangeValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case ElevationChangeUnitsFeet:
			return ElevationChangeValueFeetMinimum, ElevationChangeValueFeetMaximum
		case ElevationChangeUnitsMeters:
			return ElevationChangeValueMetersMinimum, ElevationChangeValueMetersMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
