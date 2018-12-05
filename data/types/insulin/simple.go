package insulin

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	SimpleActingTypeIntermediate = "intermediate"
	SimpleActingTypeLong         = "long"
	SimpleActingTypeRapid        = "rapid"
	SimpleActingTypeShort        = "short"
	SimpleBrandLengthMaximum     = 100
)

func SimpleActingTypes() []string {
	return []string{
		SimpleActingTypeIntermediate,
		SimpleActingTypeLong,
		SimpleActingTypeRapid,
		SimpleActingTypeShort,
	}
}

type Simple struct {
	ActingType    *string        `json:"actingType,omitempty" bson:"actingType,omitempty"`
	Brand         *string        `json:"brand,omitempty" bson:"brand,omitempty"`
	Concentration *Concentration `json:"concentration,omitempty" bson:"concentration,omitempty"`
}

func ParseSimple(parser data.ObjectParser) *Simple {
	if parser.Object() == nil {
		return nil
	}
	datum := NewSimple()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewSimple() *Simple {
	return &Simple{}
}

func (s *Simple) Parse(parser data.ObjectParser) {
	s.ActingType = parser.ParseString("actingType")
	s.Brand = parser.ParseString("brand")
	s.Concentration = ParseConcentration(parser.NewChildObjectParser("concentration"))
}

func (s *Simple) Validate(validator structure.Validator) {
	validator.String("actingType", s.ActingType).Exists().OneOf(SimpleActingTypes()...)
	validator.String("brand", s.Brand).NotEmpty().LengthLessThanOrEqualTo(SimpleBrandLengthMaximum)
	if s.Concentration != nil {
		s.Concentration.Validate(validator.WithReference("concentration"))
	}
}

func (s *Simple) Normalize(normalizer data.Normalizer) {
	if s.Concentration != nil {
		s.Concentration.Normalize(normalizer.WithReference("concentration"))
	}
}
