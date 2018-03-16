package location

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	NameLengthMaximum = 100
)

type Location struct {
	GPS  *GPS    `json:"gps,omitempty" bson:"gps,omitempty"`
	Name *string `json:"name,omitempty" bson:"name,omitempty"`
}

func ParseLocation(parser data.ObjectParser) *Location {
	if parser.Object() == nil {
		return nil
	}
	datum := NewLocation()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewLocation() *Location {
	return &Location{}
}

func (l *Location) Parse(parser data.ObjectParser) {
	l.GPS = ParseGPS(parser.NewChildObjectParser("gps"))
	l.Name = parser.ParseString("name")
}

func (l *Location) Validate(validator structure.Validator) {
	if l.GPS != nil {
		l.GPS.Validate(validator.WithReference("gps"))
	} else if l.Name == nil {
		validator.WithReference("gps").ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.String("name", l.Name).NotEmpty().LengthLessThanOrEqualTo(NameLengthMaximum)
}

func (l *Location) Normalize(normalizer data.Normalizer) {
	if l.GPS != nil {
		l.GPS.Normalize(normalizer.WithReference("gps"))
	}
}
