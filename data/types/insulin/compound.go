package insulin

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	CompoundArrayLengthMaximum = 100
	CompoundAmountMinimum      = 0.0
)

type Compound struct {
	Amount      *float64     `json:"amount,omitempty" bson:"amount,omitempty"`
	Formulation *Formulation `json:"formulation,omitempty" bson:"formulation,omitempty"`
}

func ParseCompound(parser structure.ObjectParser) *Compound {
	if !parser.Exists() {
		return nil
	}
	datum := NewCompound()
	parser.Parse(datum)
	return datum
}

func NewCompound() *Compound {
	return &Compound{}
}

func (c *Compound) Parse(parser structure.ObjectParser) {
	c.Amount = parser.Float64("amount")
	c.Formulation = ParseFormulation(parser.WithReferenceObjectParser("formulation"))
}

func (c *Compound) Validate(validator structure.Validator) {
	validator.Float64("amount", c.Amount).Exists().GreaterThanOrEqualTo(CompoundAmountMinimum)
	if c.Formulation != nil {
		c.Formulation.Validate(validator.WithReference("formulation"))
	} else {
		validator.WithReference("formulation").ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (c *Compound) Normalize(normalizer data.Normalizer) {
	if c.Formulation != nil {
		c.Formulation.Normalize(normalizer.WithReference("formulation"))
	}
}

type CompoundArray []*Compound

func ParseCompoundArray(parser structure.ArrayParser) *CompoundArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewCompoundArray()
	parser.Parse(datum)
	return datum
}

func NewCompoundArray() *CompoundArray {
	return &CompoundArray{}
}

func (c *CompoundArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*c = append(*c, ParseCompound(parser.WithReferenceObjectParser(reference)))
	}
}

func (c *CompoundArray) Validate(validator structure.Validator) {
	if length := len(*c); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	} else if length > CompoundArrayLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, CompoundArrayLengthMaximum))
	}
	for index, datum := range *c {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (c *CompoundArray) Normalize(normalizer data.Normalizer) {
	for index, datum := range *c {
		if datum != nil {
			datum.Normalize(normalizer.WithReference(strconv.Itoa(index)))
		}
	}
}
