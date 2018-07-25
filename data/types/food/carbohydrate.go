package food

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	CarbohydrateDietaryFiberGramsMaximum = 1000
	CarbohydrateDietaryFiberGramsMinimum = 0
	CarbohydrateNetGramsMaximum          = 1000
	CarbohydrateNetGramsMinimum          = 0
	CarbohydrateSugarsGramsMaximum       = 1000
	CarbohydrateSugarsGramsMinimum       = 0
	CarbohydrateTotalGramsMaximum        = 1000
	CarbohydrateTotalGramsMinimum        = 0
	CarbohydrateUnitsGrams               = "grams"
)

func CarbohydrateUnits() []string {
	return []string{
		CarbohydrateUnitsGrams,
	}
}

type Carbohydrate struct {
	DietaryFiber *int    `json:"dietaryFiber,omitempty" bson:"dietaryFiber,omitempty"`
	Net          *int    `json:"net,omitempty" bson:"net,omitempty"`
	Sugars       *int    `json:"sugars,omitempty" bson:"sugars,omitempty"`
	Total        *int    `json:"total,omitempty" bson:"total,omitempty"`
	Units        *string `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseCarbohydrate(parser data.ObjectParser) *Carbohydrate {
	if parser.Object() == nil {
		return nil
	}
	datum := NewCarbohydrate()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewCarbohydrate() *Carbohydrate {
	return &Carbohydrate{}
}

func (c *Carbohydrate) Parse(parser data.ObjectParser) {
	c.DietaryFiber = parser.ParseInteger("dietaryFiber")
	c.Net = parser.ParseInteger("net")
	c.Sugars = parser.ParseInteger("sugars")
	c.Total = parser.ParseInteger("total")
	c.Units = parser.ParseString("units")
}

func (c *Carbohydrate) Validate(validator structure.Validator) {
	validator.Int("dietaryFiber", c.DietaryFiber).InRange(CarbohydrateDietaryFiberGramsMinimum, CarbohydrateDietaryFiberGramsMaximum)
	validator.Int("net", c.Net).Exists().InRange(CarbohydrateNetGramsMinimum, CarbohydrateNetGramsMaximum)
	validator.Int("sugars", c.Sugars).InRange(CarbohydrateSugarsGramsMinimum, CarbohydrateSugarsGramsMaximum)
	validator.Int("total", c.Total).InRange(CarbohydrateTotalGramsMinimum, CarbohydrateTotalGramsMaximum)
	validator.String("units", c.Units).Exists().OneOf(CarbohydrateUnits()...)
}

func (c *Carbohydrate) Normalize(normalizer data.Normalizer) {}
