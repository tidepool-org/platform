package insulin

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type InsulinType struct {
	Formulation *Formulation `json:"formulation,omitempty" bson:"formulation,omitempty"`
	Mix         *Mix         `json:"mix,omitempty" bson:"mix,omitempty"`
}

func ParseInsulinType(parser data.ObjectParser) *InsulinType {
	if parser.Object() == nil {
		return nil
	}
	datum := NewInsulinType()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewInsulinType() *InsulinType {
	return &InsulinType{}
}

func (i *InsulinType) Parse(parser data.ObjectParser) {
	i.Formulation = ParseFormulation(parser.NewChildObjectParser("formulation"))
	i.Mix = ParseMix(parser.NewChildArrayParser("mix"))
}

func (i *InsulinType) Validate(validator structure.Validator) {
	if i.Formulation != nil {
		i.Formulation.Validate(validator.WithReference("formulation"))
		if i.Mix != nil {
			validator.WithReference("mix").ReportError(structureValidator.ErrorValueExists())
		}
	} else if i.Mix != nil {
		i.Mix.Validate(validator.WithReference("mix"))
	} else {
		validator.WithReference("formulation").ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (i *InsulinType) Normalize(normalizer data.Normalizer) {
	if i.Formulation != nil {
		i.Formulation.Normalize(normalizer.WithReference("formulation"))
	}
	if i.Mix != nil {
		i.Mix.Normalize(normalizer.WithReference("mix"))
	}
}
