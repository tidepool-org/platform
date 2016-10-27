package pump

import "github.com/tidepool-org/platform/data"

type CarbohydrateRatio struct {
	Amount *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
	Start  *int     `json:"start,omitempty" bson:"start,omitempty"`
}

func NewCarbohydrateRatio() *CarbohydrateRatio {
	return &CarbohydrateRatio{}
}

func (c *CarbohydrateRatio) Parse(parser data.ObjectParser) {
	c.Amount = parser.ParseFloat("amount")
	c.Start = parser.ParseInteger("start")
}

func (c *CarbohydrateRatio) Validate(validator data.Validator) {
	validator.ValidateFloat("amount", c.Amount).Exists().InRange(0.0, 250.0)
	validator.ValidateInteger("start", c.Start).Exists().InRange(0, 86400000)
}

func (c *CarbohydrateRatio) Normalize(normalizer data.Normalizer) {
}

func ParseCarbohydrateRatio(parser data.ObjectParser) *CarbohydrateRatio {
	var carbohydrateRatio *CarbohydrateRatio
	if parser.Object() != nil {
		carbohydrateRatio = NewCarbohydrateRatio()
		carbohydrateRatio.Parse(parser)
		parser.ProcessNotParsed()
	}
	return carbohydrateRatio
}

func ParseCarbohydrateRatioArray(parser data.ArrayParser) *[]*CarbohydrateRatio {
	var carbohydrateRatioArray *[]*CarbohydrateRatio
	if parser.Array() != nil {
		carbohydrateRatioArray = &[]*CarbohydrateRatio{}
		for index := range *parser.Array() {
			*carbohydrateRatioArray = append(*carbohydrateRatioArray, ParseCarbohydrateRatio(parser.NewChildObjectParser(index)))
		}
		parser.ProcessNotParsed()
	}
	return carbohydrateRatioArray
}
