package location

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	GPSFloorMaximum = 1000
	GPSFloorMinimum = -1000
)

type GPS struct {
	Elevation          *Elevation `json:"elevation,omitempty" bson:"elevation,omitempty"`
	Floor              *int       `json:"floor,omitempty" bson:"floor,omitempty"`
	HorizontalAccuracy *Accuracy  `json:"horizontalAccuracy,omitempty" bson:"horizontalAccuracy,omitempty"`
	Latitude           *Latitude  `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude          *Longitude `json:"longitude,omitempty" bson:"longitude,omitempty"`
	VerticalAccuracy   *Accuracy  `json:"verticalAccuracy,omitempty" bson:"verticalAccuracy,omitempty"`
}

func ParseGPS(parser data.ObjectParser) *GPS {
	if parser.Object() == nil {
		return nil
	}
	datum := NewGPS()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewGPS() *GPS {
	return &GPS{}
}

func (g *GPS) Parse(parser data.ObjectParser) {
	g.Elevation = ParseElevation(parser.NewChildObjectParser("elevation"))
	g.Floor = parser.ParseInteger("floor")
	g.HorizontalAccuracy = ParseAccuracy(parser.NewChildObjectParser("horizontalAccuracy"))
	g.Latitude = ParseLatitude(parser.NewChildObjectParser("latitude"))
	g.Longitude = ParseLongitude(parser.NewChildObjectParser("longitude"))
	g.VerticalAccuracy = ParseAccuracy(parser.NewChildObjectParser("verticalAccuracy"))
}

func (g *GPS) Validate(validator structure.Validator) {
	if g.Elevation != nil {
		g.Elevation.Validate(validator.WithReference("elevation"))
	}
	validator.Int("floor", g.Floor).InRange(GPSFloorMinimum, GPSFloorMaximum)
	if g.HorizontalAccuracy != nil {
		g.HorizontalAccuracy.Validate(validator.WithReference("horizontalAccuracy"))
	}
	if g.Latitude != nil {
		g.Latitude.Validate(validator.WithReference("latitude"))
	} else {
		validator.WithReference("latitude").ReportError(structureValidator.ErrorValueNotExists())
	}
	if g.Longitude != nil {
		g.Longitude.Validate(validator.WithReference("longitude"))
	} else {
		validator.WithReference("longitude").ReportError(structureValidator.ErrorValueNotExists())
	}
	if g.VerticalAccuracy != nil {
		g.VerticalAccuracy.Validate(validator.WithReference("verticalAccuracy"))
	}
}

func (g *GPS) Normalize(normalizer data.Normalizer) {
	if g.Elevation != nil {
		g.Elevation.Normalize(normalizer.WithReference("elevation"))
	}
	if g.HorizontalAccuracy != nil {
		g.HorizontalAccuracy.Normalize(normalizer.WithReference("horizontalAccuracy"))
	}
	if g.Latitude != nil {
		g.Latitude.Normalize(normalizer.WithReference("latitude"))
	}
	if g.Longitude != nil {
		g.Longitude.Normalize(normalizer.WithReference("longitude"))
	}
	if g.VerticalAccuracy != nil {
		g.VerticalAccuracy.Normalize(normalizer.WithReference("verticalAccuracy"))
	}
}
