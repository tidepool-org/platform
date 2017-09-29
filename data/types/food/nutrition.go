package food

import "github.com/tidepool-org/platform/data"

type Nutrition struct {
	Carbohydrates *Carbohydrates `json:"carbohydrates,omitempty" bson:"carbohydrates,omitempty"`
}

func NewNutrition() *Nutrition {
	return &Nutrition{}
}

func (n *Nutrition) Parse(parser data.ObjectParser) {
	n.Carbohydrates = ParseCarbohydrates(parser.NewChildObjectParser("carbohydrates"))
}

func (n *Nutrition) Validate(validator data.Validator) {
	if n.Carbohydrates != nil {
		n.Carbohydrates.Validate(validator.NewChildValidator("carbohydrates"))
	}
}

func (n *Nutrition) Normalize(normalizer data.Normalizer) {
	if n.Carbohydrates != nil {
		n.Carbohydrates.Normalize(normalizer.NewChildNormalizer("carbohydrates"))
	}
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
