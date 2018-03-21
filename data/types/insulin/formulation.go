package insulin

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	FormulationActingTypeIntermediate = "intermediate"
	FormulationActingTypeLong         = "long"
	FormulationActingTypeRapid        = "rapid"
	FormulationActingTypeShort        = "short"
	FormulationBrandLengthMaximum     = 100
	FormulationNameLengthMaximum      = 100
)

func FormulationActingTypes() []string {
	return []string{
		FormulationActingTypeIntermediate,
		FormulationActingTypeLong,
		FormulationActingTypeRapid,
		FormulationActingTypeShort,
	}
}

type Formulation struct {
	ActingType    *string        `json:"actingType,omitempty" bson:"actingType,omitempty"`
	Brand         *string        `json:"brand,omitempty" bson:"brand,omitempty"`
	Concentration *Concentration `json:"concentration,omitempty" bson:"concentration,omitempty"`
	Name          *string        `json:"name,omitempty" bson:"name,omitempty"`
}

func ParseFormulation(parser data.ObjectParser) *Formulation {
	if parser.Object() == nil {
		return nil
	}
	datum := NewFormulation()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewFormulation() *Formulation {
	return &Formulation{}
}

func (f *Formulation) Parse(parser data.ObjectParser) {
	f.ActingType = parser.ParseString("actingType")
	f.Brand = parser.ParseString("brand")
	f.Concentration = ParseConcentration(parser.NewChildObjectParser("concentration"))
	f.Name = parser.ParseString("name")
}

func (f *Formulation) Validate(validator structure.Validator) {
	validator.String("actingType", f.ActingType).OneOf(FormulationActingTypes()...)
	validator.String("brand", f.Brand).NotEmpty().LengthLessThanOrEqualTo(FormulationBrandLengthMaximum)
	if f.Concentration != nil {
		f.Concentration.Validate(validator.WithReference("concentration"))
	}
	validator.String("name", f.Name).NotEmpty().LengthLessThanOrEqualTo(FormulationNameLengthMaximum)
}

func (f *Formulation) Normalize(normalizer data.Normalizer) {
	if f.Concentration != nil {
		f.Concentration.Normalize(normalizer.WithReference("concentration"))
	}
}
