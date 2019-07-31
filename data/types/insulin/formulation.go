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

func ParseFormulation(parser structure.ObjectParser) *Formulation {
	if !parser.Exists() {
		return nil
	}
	datum := NewFormulation()
	parser.Parse(datum)
	return datum
}

func NewFormulation() *Formulation {
	return &Formulation{}
}

func (f *Formulation) Parse(parser structure.ObjectParser) {
	f.Compounds = ParseCompoundArray(parser.WithReferenceArrayParser("compounds"))
	f.Name = parser.String("name")
	f.Simple = ParseSimple(parser.WithReferenceObjectParser("simple"))
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
