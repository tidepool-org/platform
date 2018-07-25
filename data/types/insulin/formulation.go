package insulin

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	FormulationNameLengthMaximum = 100
)

type Formulation struct {
	Compounds *CompoundArray `json:"compounds,omitempty" bson:"compounds,omitempty"`
	Name      *string        `json:"name,omitempty" bson:"name,omitempty"`
	Simple    *Simple        `json:"simple,omitempty" bson:"simple,omitempty"`
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
	f.Compounds = ParseCompoundArray(parser.NewChildArrayParser("compounds"))
	f.Name = parser.ParseString("name")
	f.Simple = ParseSimple(parser.NewChildObjectParser("simple"))
}

func (f *Formulation) Validate(validator structure.Validator) {
	if f.Compounds != nil {
		if f.Simple != nil {
			validator.WithReference("compounds").ReportError(structureValidator.ErrorValueExists())
		} else {
			f.Compounds.Validate(validator.WithReference("compounds"))
		}
	}
	validator.String("name", f.Name).NotEmpty().LengthLessThanOrEqualTo(FormulationNameLengthMaximum)
	if f.Simple != nil {
		f.Simple.Validate(validator.WithReference("simple"))
	} else if f.Compounds == nil && f.Name == nil {
		validator.WithReference("simple").ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (f *Formulation) Normalize(normalizer data.Normalizer) {
	if f.Compounds != nil {
		f.Compounds.Normalize(normalizer.WithReference("compounds"))
	}
	if f.Simple != nil {
		f.Simple.Normalize(normalizer.WithReference("simple"))
	}
}
