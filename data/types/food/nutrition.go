package food

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type Nutrition struct {
	Carbohydrate *Carbohydrate `json:"carbohydrate,omitempty" bson:"carbohydrate,omitempty"`
	Energy       *Energy       `json:"energy,omitempty" bson:"energy,omitempty"`
	Fat          *Fat          `json:"fat,omitempty" bson:"fat,omitempty"`
	Protein      *Protein      `json:"protein,omitempty" bson:"protein,omitempty"`
}

func ParseNutrition(parser data.ObjectParser) *Nutrition {
	if parser.Object() == nil {
		return nil
	}
	datum := NewNutrition()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewNutrition() *Nutrition {
	return &Nutrition{}
}

func (n *Nutrition) Parse(parser data.ObjectParser) {
	n.Carbohydrate = ParseCarbohydrate(parser.NewChildObjectParser("carbohydrate"))
	n.Energy = ParseEnergy(parser.NewChildObjectParser("energy"))
	n.Fat = ParseFat(parser.NewChildObjectParser("fat"))
	n.Protein = ParseProtein(parser.NewChildObjectParser("protein"))
}

func (n *Nutrition) Validate(validator structure.Validator) {
	if n.Carbohydrate != nil {
		n.Carbohydrate.Validate(validator.WithReference("carbohydrate"))
	}
	if n.Energy != nil {
		n.Energy.Validate(validator.WithReference("energy"))
	}
	if n.Fat != nil {
		n.Fat.Validate(validator.WithReference("fat"))
	}
	if n.Protein != nil {
		n.Protein.Validate(validator.WithReference("protein"))
	}
}

func (n *Nutrition) Normalize(normalizer data.Normalizer) {
	if n.Carbohydrate != nil {
		n.Carbohydrate.Normalize(normalizer.WithReference("carbohydrate"))
	}
	if n.Energy != nil {
		n.Energy.Normalize(normalizer.WithReference("energy"))
	}
	if n.Fat != nil {
		n.Fat.Normalize(normalizer.WithReference("fat"))
	}
	if n.Protein != nil {
		n.Protein.Normalize(normalizer.WithReference("protein"))
	}
}
