package food

import "github.com/tidepool-org/platform/data"

const (
	UnitsGrams = "grams"

	ValueGramsMinimum = 0
	ValueGramsMaximum = 1000
)

type Carbohydrates struct {
	Net   *int    `json:"net,omitempty" bson:"net,omitempty"`
	Units *string `json:"units,omitempty" bson:"units,omitempty"`
}

func NewCarbohydrates() *Carbohydrates {
	return &Carbohydrates{}
}

func (c *Carbohydrates) Parse(parser data.ObjectParser) {
	c.Net = parser.ParseInteger("net")
	c.Units = parser.ParseString("units")
}

func (c *Carbohydrates) Validate(validator data.Validator) {
	validator.ValidateInteger("net", c.Net).Exists().InRange(ValueGramsMinimum, ValueGramsMaximum)
	validator.ValidateString("units", c.Units).Exists().EqualTo(UnitsGrams)
}

func (c *Carbohydrates) Normalize(normalizer data.Normalizer) {
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
