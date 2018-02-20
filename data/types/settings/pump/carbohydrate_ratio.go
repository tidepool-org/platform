package pump

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	CarbohydrateRatioAmountMaximum = 250.0
	CarbohydrateRatioAmountMinimum = 0.0
	CarbohydrateRatioStartMaximum  = 86400000
	CarbohydrateRatioStartMinimum  = 0
)

type CarbohydrateRatio struct {
	Amount *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
	Start  *int     `json:"start,omitempty" bson:"start,omitempty"`
}

func ParseCarbohydrateRatio(parser data.ObjectParser) *CarbohydrateRatio {
	if parser.Object() == nil {
		return nil
	}
	carbohydrateRatio := NewCarbohydrateRatio()
	carbohydrateRatio.Parse(parser)
	parser.ProcessNotParsed()
	return carbohydrateRatio
}

func NewCarbohydrateRatio() *CarbohydrateRatio {
	return &CarbohydrateRatio{}
}

func (c *CarbohydrateRatio) Parse(parser data.ObjectParser) {
	c.Amount = parser.ParseFloat("amount")
	c.Start = parser.ParseInteger("start")
}

func (c *CarbohydrateRatio) Validate(validator structure.Validator) {
	validator.Float64("amount", c.Amount).Exists().InRange(CarbohydrateRatioAmountMinimum, CarbohydrateRatioAmountMaximum)
	validator.Int("start", c.Start).Exists().InRange(CarbohydrateRatioStartMinimum, CarbohydrateRatioStartMaximum)
}

func (c *CarbohydrateRatio) Normalize(normalizer data.Normalizer) {}

// TODO: Can/should we validate that each Start in the array is greater than the previous Start?

type CarbohydrateRatioArray []*CarbohydrateRatio

func ParseCarbohydrateRatioArray(parser data.ArrayParser) *CarbohydrateRatioArray {
	if parser.Array() == nil {
		return nil
	}
	carbohydrateRatioArray := NewCarbohydrateRatioArray()
	carbohydrateRatioArray.Parse(parser)
	parser.ProcessNotParsed()
	return carbohydrateRatioArray
}

func NewCarbohydrateRatioArray() *CarbohydrateRatioArray {
	return &CarbohydrateRatioArray{}
}

func (c *CarbohydrateRatioArray) Parse(parser data.ArrayParser) {
	for index := range *parser.Array() {
		*c = append(*c, ParseCarbohydrateRatio(parser.NewChildObjectParser(index)))
	}
}

func (c *CarbohydrateRatioArray) Validate(validator structure.Validator) {
	for index, carbohydrateRatio := range *c {
		carbohydrateRatioValidator := validator.WithReference(strconv.Itoa(index))
		if carbohydrateRatio != nil {
			carbohydrateRatio.Validate(carbohydrateRatioValidator)
		} else {
			carbohydrateRatioValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (c *CarbohydrateRatioArray) Normalize(normalizer data.Normalizer) {
	for index, carbohydrateRatio := range *c {
		if carbohydrateRatio != nil {
			carbohydrateRatio.Normalize(normalizer.WithReference(strconv.Itoa(index)))
		}
	}
}
