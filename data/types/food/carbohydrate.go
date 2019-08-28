package food

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	CarbohydrateDietaryFiberGramsMaximum = 1000.0
	CarbohydrateDietaryFiberGramsMinimum = 0.0
	CarbohydrateNetGramsMaximum          = 1000.0
	CarbohydrateNetGramsMinimum          = 0.0
	CarbohydrateSugarsGramsMaximum       = 1000.0
	CarbohydrateSugarsGramsMinimum       = 0.0
	CarbohydrateTotalGramsMaximum        = 1000.0
	CarbohydrateTotalGramsMinimum        = 0.0
	CarbohydrateUnitsGrams               = "grams"
)

func CarbohydrateUnits() []string {
	return []string{
		CarbohydrateUnitsGrams,
	}
}

type Carbohydrate struct {
	DietaryFiber       *float64 `json:"dietaryFiber,omitempty" bson:"dietaryFiber,omitempty"`
	Net                *float64 `json:"net,omitempty" bson:"net,omitempty"`
	Sugars             *float64 `json:"sugars,omitempty" bson:"sugars,omitempty"`
	Total              *float64 `json:"total,omitempty" bson:"total,omitempty"`
	Units              *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseCarbohydrate(parser structure.ObjectParser) *Carbohydrate {
	if !parser.Exists() {
		return nil
	}
	datum := NewCarbohydrate()
	parser.Parse(datum)
	return datum
}

func NewCarbohydrate() *Carbohydrate {
	return &Carbohydrate{}
}

func (c *Carbohydrate) Parse(parser structure.ObjectParser) {
	c.DietaryFiber = parser.Float64("dietaryFiber")
	c.Net = parser.Float64("net")
	c.Sugars = parser.Float64("sugars")
	c.Total = parser.Float64("total")
	c.Units = parser.String("units")
}

func (c *Carbohydrate) Validate(validator structure.Validator) {
	validator.Float64("dietaryFiber", c.DietaryFiber).InRange(CarbohydrateDietaryFiberGramsMinimum, CarbohydrateDietaryFiberGramsMaximum)
	validator.Float64("net", c.Net).Exists().InRange(CarbohydrateNetGramsMinimum, CarbohydrateNetGramsMaximum)
	validator.Float64("sugars", c.Sugars).InRange(CarbohydrateSugarsGramsMinimum, CarbohydrateSugarsGramsMaximum)
	validator.Float64("total", c.Total).InRange(CarbohydrateTotalGramsMinimum, CarbohydrateTotalGramsMaximum)
	validator.String("units", c.Units).Exists().OneOf(CarbohydrateUnits()...)
}

func (c *Carbohydrate) Normalize(normalizer data.Normalizer) {}
