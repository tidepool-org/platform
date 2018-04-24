package food

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	UnitsGrams        = "grams"
	ValueGramsMaximum = 1000
	ValueGramsMinimum = 0
)

func Units() []string {
	return []string{
		UnitsGrams,
	}
}

type Carbohydrates struct {
	Net   *int    `json:"net,omitempty" bson:"net,omitempty"`
	Units *string `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseCarbohydrates(parser data.ObjectParser) *Carbohydrates {
	if parser.Object() == nil {
		return nil
	}
	carbohydrates := NewCarbohydrates()
	carbohydrates.Parse(parser)
	parser.ProcessNotParsed()
	return carbohydrates
}

func NewCarbohydrates() *Carbohydrates {
	return &Carbohydrates{}
}

func (c *Carbohydrates) Parse(parser data.ObjectParser) {
	c.Net = parser.ParseInteger("net")
	c.Units = parser.ParseString("units")
}

func (c *Carbohydrates) Validate(validator structure.Validator) {
	validator.Int("net", c.Net).Exists().InRange(ValueGramsMinimum, ValueGramsMaximum)
	validator.String("units", c.Units).Exists().OneOf(Units()...)
}

func (c *Carbohydrates) Normalize(normalizer data.Normalizer) {}
