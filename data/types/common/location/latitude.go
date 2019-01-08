package location

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	LatitudeUnitsDegrees = "degrees"
	LatitudeValueMaximum = 90.0
	LatitudeValueMinimum = -90.0
)

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
	validator.String("units", l.Units).Exists().EqualTo(LatitudeUnitsDegrees)
	validator.Float64("value", l.Value).Exists().InRange(LatitudeValueMinimum, LatitudeValueMaximum)
}

func (l *Latitude) Normalize(normalizer data.Normalizer) {}
