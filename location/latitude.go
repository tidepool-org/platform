package location

import (
	"math"

	"github.com/tidepool-org/platform/structure"
)

const (
	LatitudeUnitsDegrees        = "degrees"
	LatitudeValueDegreesMaximum = 90.0
	LatitudeValueDegreesMinimum = -90.0
)

func LatitudeUnits() []string {
	return []string{
		LatitudeUnitsDegrees,
	}
}

type Latitude struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseLatitude(parser structure.ObjectParser) *Latitude {
	if !parser.Exists() {
		return nil
	}
	datum := NewLatitude()
	parser.Parse(datum)
	return datum
}

func NewLatitude() *Latitude {
	return &Latitude{}
}

func (l *Latitude) Parse(parser structure.ObjectParser) {
	l.Units = parser.String("units")
	l.Value = parser.Float64("value")
}

func (l *Latitude) Validate(validator structure.Validator) {
	validator.String("units", l.Units).Exists().OneOf(LatitudeUnits()...)
	validator.Float64("value", l.Value).Exists().InRange(LatitudeValueRangeForUnits(l.Units))
}

func LatitudeValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case LatitudeUnitsDegrees:
			return LatitudeValueDegreesMinimum, LatitudeValueDegreesMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
