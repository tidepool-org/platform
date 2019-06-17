package location

import (
	"math"

	"github.com/tidepool-org/platform/structure"
)

const (
	LongitudeUnitsDegrees        = "degrees"
	LongitudeValueDegreesMaximum = 180.0
	LongitudeValueDegreesMinimum = -180.0
)

func LongitudeUnits() []string {
	return []string{
		LongitudeUnitsDegrees,
	}
}

type Longitude struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseLongitude(parser structure.ObjectParser) *Longitude {
	if !parser.Exists() {
		return nil
	}
	datum := NewLongitude()
	parser.Parse(datum)
	return datum
}

func NewLongitude() *Longitude {
	return &Longitude{}
}

func (l *Longitude) Parse(parser structure.ObjectParser) {
	l.Units = parser.String("units")
	l.Value = parser.Float64("value")
}

func (l *Longitude) Validate(validator structure.Validator) {
	validator.String("units", l.Units).Exists().OneOf(LongitudeUnits()...)
	validator.Float64("value", l.Value).Exists().InRange(LongitudeValueRangeForUnits(l.Units))
}

func LongitudeValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case LongitudeUnitsDegrees:
			return LongitudeValueDegreesMinimum, LongitudeValueDegreesMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
