package insulin

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	MixElementAmountMinimum = 0.0
)

type MixElement struct {
	Amount      *float64     `json:"amount,omitempty" bson:"amount,omitempty"`
	Formulation *Formulation `json:"formulation,omitempty" bson:"formulation,omitempty"`
}

func ParseMixElement(parser data.ObjectParser) *MixElement {
	if parser.Object() == nil {
		return nil
	}
	datum := NewMixElement()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewMixElement() *MixElement {
	return &MixElement{}
}

func (m *MixElement) Parse(parser data.ObjectParser) {
	m.Amount = parser.ParseFloat("amount")
	m.Formulation = ParseFormulation(parser.NewChildObjectParser("formulation"))
}

func (m *MixElement) Validate(validator structure.Validator) {
	validator.Float64("amount", m.Amount).Exists().GreaterThanOrEqualTo(MixElementAmountMinimum)
	if m.Formulation != nil {
		m.Formulation.Validate(validator.WithReference("formulation"))
	} else {
		validator.WithReference("formulation").ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (m *MixElement) Normalize(normalizer data.Normalizer) {
	if m.Formulation != nil {
		m.Formulation.Normalize(normalizer.WithReference("formulation"))
	}
}

type Mix []*MixElement

func ParseMix(parser data.ArrayParser) *Mix {
	if parser.Array() == nil {
		return nil
	}
	datum := NewMix()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewMix() *Mix {
	return &Mix{}
}

func (m *Mix) Parse(parser data.ArrayParser) {
	for index := range *parser.Array() {
		*m = append(*m, ParseMixElement(parser.NewChildObjectParser(index)))
	}
}

func (m *Mix) Validate(validator structure.Validator) {
	for index, datum := range *m {
		datumValidator := validator.WithReference(strconv.Itoa(index))
		if datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (m *Mix) Normalize(normalizer data.Normalizer) {
	for index, datum := range *m {
		if datum != nil {
			datum.Normalize(normalizer.WithReference(strconv.Itoa(index)))
		}
	}
}
