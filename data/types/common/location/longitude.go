package location

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	LongitudeUnitsDegrees = "degrees"
	LongitudeValueMaximum = 180.0
	LongitudeValueMinimum = -180.0
)

type Longitude struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseLongitude(parser data.ObjectParser) *Longitude {
	if parser.Object() == nil {
		return nil
	}
	datum := NewLongitude()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewLongitude() *Longitude {
	return &Longitude{}
}

func (l *Longitude) Parse(parser data.ObjectParser) {
	l.Units = parser.ParseString("units")
	l.Value = parser.ParseFloat("value")
}

func (l *Longitude) Validate(validator structure.Validator) {
	validator.String("units", l.Units).Exists().EqualTo(LongitudeUnitsDegrees)
	validator.Float64("value", l.Value).Exists().InRange(LongitudeValueMinimum, LongitudeValueMaximum)
}

func (l *Longitude) Normalize(normalizer data.Normalizer) {}
