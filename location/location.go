package location

import (
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const NameLengthMaximum = 100

type Location struct {
	GPS  *GPS    `json:"gps,omitempty" bson:"gps,omitempty"`
	Name *string `json:"name,omitempty" bson:"name,omitempty"`
}

func ParseLocation(parser structure.ObjectParser) *Location {
	if !parser.Exists() {
		return nil
	}
	datum := NewLocation()
	parser.Parse(datum)
	return datum
}

func NewLocation() *Location {
	return &Location{}
}

func (l *Location) Parse(parser structure.ObjectParser) {
	l.GPS = ParseGPS(parser.WithReferenceObjectParser("gps"))
	l.Name = parser.String("name")
}

func (l *Location) Validate(validator structure.Validator) {
	if l.GPS != nil {
		l.GPS.Validate(validator.WithReference("gps"))
	} else if l.Name == nil {
		validator.ReportError(structureValidator.ErrorValuesNotExistForAny("gps", "name"))
	}
	validator.String("name", l.Name).NotEmpty().LengthLessThanOrEqualTo(NameLengthMaximum)
}
