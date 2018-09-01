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

func ParseCompound(parser data.ObjectParser) *Compound {
	if parser.Object() == nil {
		return nil
	}
	datum := NewCompound()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewCompound() *Compound {
	return &Compound{}
}

func (c *Compound) Parse(parser data.ObjectParser) {
	c.Amount = parser.ParseFloat("amount")
	c.Formulation = ParseFormulation(parser.NewChildObjectParser("formulation"))
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

func ParseCompoundArray(parser data.ArrayParser) *CompoundArray {
	if parser.Array() == nil {
		return nil
	}
	datum := NewCompoundArray()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewCompoundArray() *CompoundArray {
	return &CompoundArray{}
}

func (c *CompoundArray) Parse(parser data.ArrayParser) {
	for index := range *parser.Array() {
		*c = append(*c, ParseCompound(parser.NewChildObjectParser(index)))
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
