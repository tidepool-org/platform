package location

import (
	"github.com/tidepool-org/platform/data"
	dataTypesCommonOrigin "github.com/tidepool-org/platform/data/types/common/origin"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	GPSFloorMaximum = 1000
	GPSFloorMinimum = -1000
)

type GPS struct {
	Elevation          *Elevation                    `json:"elevation,omitempty" bson:"elevation,omitempty"`
	Floor              *int                          `json:"floor,omitempty" bson:"floor,omitempty"`
	HorizontalAccuracy *Accuracy                     `json:"horizontalAccuracy,omitempty" bson:"horizontalAccuracy,omitempty"`
	Latitude           *Latitude                     `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude          *Longitude                    `json:"longitude,omitempty" bson:"longitude,omitempty"`
	Origin             *dataTypesCommonOrigin.Origin `json:"origin,omitempty" bson:"origin,omitempty"`
	VerticalAccuracy   *Accuracy                     `json:"verticalAccuracy,omitempty" bson:"verticalAccuracy,omitempty"`
}

func ParseGPS(parser structure.ObjectParser) *GPS {
	if !parser.Exists() {
		return nil
	}
	datum := NewGPS()
	parser.Parse(datum)
	return datum
}

func NewGPS() *GPS {
	return &GPS{}
}

func (g *GPS) Parse(parser structure.ObjectParser) {
	g.Elevation = ParseElevation(parser.WithReferenceObjectParser("elevation"))
	g.Floor = parser.Int("floor")
	g.HorizontalAccuracy = ParseAccuracy(parser.WithReferenceObjectParser("horizontalAccuracy"))
	g.Latitude = ParseLatitude(parser.WithReferenceObjectParser("latitude"))
	g.Longitude = ParseLongitude(parser.WithReferenceObjectParser("longitude"))
	g.Origin = dataTypesCommonOrigin.ParseOrigin(parser.WithReferenceObjectParser("origin"))
	g.VerticalAccuracy = ParseAccuracy(parser.WithReferenceObjectParser("verticalAccuracy"))
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
	if g.Origin != nil {
		g.Origin.Validate(validator.WithReference("origin"))
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
	if g.Origin != nil {
		g.Origin.Normalize(normalizer.WithReference("origin"))
	}
	if g.VerticalAccuracy != nil {
		g.VerticalAccuracy.Normalize(normalizer.WithReference("verticalAccuracy"))
	}
}
