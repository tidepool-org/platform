package location

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	ElevationUnitsFeet          = "feet"
	ElevationUnitsMeter         = "meters"
	ElevationValueFeetMaximum   = 10000.0 / 0.3048
	ElevationValueFeetMinimum   = 0.0
	ElevationValueMetersMaximum = 10000.0
	ElevationValueMetersMinimum = 0.0
)

func ElevationUnits() []string {
	return []string{
		ElevationUnitsFeet,
		ElevationUnitsMeter,
	}
}

type Elevation struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseElevation(parser data.ObjectParser) *Elevation {
	if parser.Object() == nil {
		return nil
	}
	datum := NewElevation()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewElevation() *Elevation {
	return &Elevation{}
}

func (e *Elevation) Parse(parser data.ObjectParser) {
	e.Units = parser.ParseString("units")
	e.Value = parser.ParseFloat("value")
}

func (e *Elevation) Validate(validator structure.Validator) {
	validator.String("units", e.Units).Exists().OneOf(ElevationUnits()...)
	validator.Float64("value", e.Value).Exists().InRange(ElevationValueRangeForUnits(e.Units))
}

func (e *Elevation) Normalize(normalizer data.Normalizer) {}

func ElevationValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case ElevationUnitsFeet:
			return ElevationValueFeetMinimum, ElevationValueFeetMaximum
		case ElevationUnitsMeter:
			return ElevationValueMetersMinimum, ElevationValueMetersMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
