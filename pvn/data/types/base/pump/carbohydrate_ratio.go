package pump

import "github.com/tidepool-org/platform/pvn/data"

type CarbohydrateRatio struct {
	Amount *int `json:"amount" bson:"amount"`
	Start  *int `json:"start" bson:"start"`
}

func NewCarbohydrateRatio() *CarbohydrateRatio {
	return &CarbohydrateRatio{}
}

func (c *CarbohydrateRatio) Parse(parser data.ObjectParser) {
	c.Amount = parser.ParseInteger("amount")
	c.Start = parser.ParseInteger("start")
}

func (c *CarbohydrateRatio) Validate(validator data.Validator) {
	validator.ValidateInteger("amount", c.Amount).Exists().InRange(0, 250)
	validator.ValidateInteger("start", c.Start).Exists().InRange(0, 86400000)
}

func (c *CarbohydrateRatio) Normalize(normalizer data.Normalizer) {
}

func ParseCarbohydrateRatio(parser data.ObjectParser) *CarbohydrateRatio {
	var carbohydrateRatio *CarbohydrateRatio

	if parser.Object() != nil {
		carbohydrateRatio = NewCarbohydrateRatio()
		carbohydrateRatio.Parse(parser)
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
	}
	return carbohydrateRatioArray
}
