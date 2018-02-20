package food

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type Nutrition struct {
	Carbohydrates *Carbohydrates `json:"carbohydrates,omitempty" bson:"carbohydrates,omitempty"`
}

func ParseNutrition(parser data.ObjectParser) *Nutrition {
	if parser.Object() == nil {
		return nil
	}
	nutrition := NewNutrition()
	nutrition.Parse(parser)
	parser.ProcessNotParsed()
	return nutrition
}

func NewNutrition() *Nutrition {
	return &Nutrition{}
}

func (n *Nutrition) Parse(parser data.ObjectParser) {
	n.Carbohydrates = ParseCarbohydrates(parser.NewChildObjectParser("carbohydrates"))
}

func (n *Nutrition) Validate(validator structure.Validator) {
	if n.Carbohydrates != nil {
		n.Carbohydrates.Validate(validator.WithReference("carbohydrates"))
	}
}

func (n *Nutrition) Normalize(normalizer data.Normalizer) {
	if n.Carbohydrates != nil {
		n.Carbohydrates.Normalize(normalizer.WithReference("carbohydrates"))
	}
}
