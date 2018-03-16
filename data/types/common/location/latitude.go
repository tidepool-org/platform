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

func ParseLatitude(parser data.ObjectParser) *Latitude {
	if parser.Object() == nil {
		return nil
	}
	datum := NewLatitude()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewLatitude() *Latitude {
	return &Latitude{}
}

func (l *Latitude) Parse(parser data.ObjectParser) {
	l.Units = parser.ParseString("units")
	l.Value = parser.ParseFloat("value")
}

func (l *Latitude) Validate(validator structure.Validator) {
	validator.String("units", l.Units).Exists().EqualTo(LatitudeUnitsDegrees)
	validator.Float64("value", l.Value).Exists().InRange(LatitudeValueMinimum, LatitudeValueMaximum)
}

func (l *Latitude) Normalize(normalizer data.Normalizer) {}
