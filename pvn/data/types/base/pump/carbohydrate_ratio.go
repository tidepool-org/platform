package pump

import (
	"fmt"

	"github.com/tidepool-org/platform/pvn/data"
)

type CarbohydrateRatio struct {
	Amount *float64 `json:"amount" bson:"amount"`
	Start  *int     `json:"start" bson:"start"`
}

func NewCarbohydrateRatio() *CarbohydrateRatio {
	return &CarbohydrateRatio{}
}

func (c *CarbohydrateRatio) Parse(parser data.ObjectParser) {
	c.Amount = parser.ParseFloat("amount")
	c.Start = parser.ParseInteger("start")
}

func (c *CarbohydrateRatio) Validate(validator data.Validator) {
	validator.ValidateFloat("amount", c.Amount).Exists().GreaterThanOrEqualTo(0)
	validator.ValidateInteger("start", c.Start).Exists().GreaterThanOrEqualTo(0)
}

func (b *CarbohydrateRatio) Normalize(normalizer data.Normalizer) {
}

func ParseCarbohydrateRatio(parser data.ObjectParser) *CarbohydrateRatio {
	var carbohydrateRatio *CarbohydrateRatio

	fmt.Println("ParseCarbohydrateRatio", parser.Object())

	if parser.Object() != nil {
		carbohydrateRatio = NewCarbohydrateRatio()
		carbohydrateRatio.Parse(parser)
	}
	return carbohydrateRatio
}

func ParseCarbohydrateRatioArray(parser data.ArrayParser) *[]*CarbohydrateRatio {

	fmt.Println("ParseCarbohydrateRatioArray", parser.Array())

	var carbohydrateRatioArray *[]*CarbohydrateRatio
	if parser.Array() != nil {
		carbohydrateRatioArray = &[]*CarbohydrateRatio{}
		for index := range *parser.Array() {
			*carbohydrateRatioArray = append(*carbohydrateRatioArray, ParseCarbohydrateRatio(parser.NewChildObjectParser(index)))
		}
	}
	return carbohydrateRatioArray
}
