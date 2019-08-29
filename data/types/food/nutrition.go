package food

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	AbsorptionDurationMaximum = 86400
	AbsorptionDurationMinimum = 0
)

type Nutrition struct {
	AbsorptionDuration *int          `json:"absorptionDuration,omitempty" bson:"absorptionDuration,omitempty"`
	Carbohydrate       *Carbohydrate `json:"carbohydrate,omitempty" bson:"carbohydrate,omitempty"`
	Energy             *Energy       `json:"energy,omitempty" bson:"energy,omitempty"`
	Fat                *Fat          `json:"fat,omitempty" bson:"fat,omitempty"`
	Protein            *Protein      `json:"protein,omitempty" bson:"protein,omitempty"`
}

func ParseNutrition(parser structure.ObjectParser) *Nutrition {
	if !parser.Exists() {
		return nil
	}
	datum := NewNutrition()
	parser.Parse(datum)
	return datum
}

func NewNutrition() *Nutrition {
	return &Nutrition{}
}

func (n *Nutrition) Parse(parser structure.ObjectParser) {
	n.AbsorptionDuration = parser.Int("absorptionDuration")
	n.Carbohydrate = ParseCarbohydrate(parser.WithReferenceObjectParser("carbohydrate"))
	n.Energy = ParseEnergy(parser.WithReferenceObjectParser("energy"))
	n.Fat = ParseFat(parser.WithReferenceObjectParser("fat"))
	n.Protein = ParseProtein(parser.WithReferenceObjectParser("protein"))
}

func (n *Nutrition) Validate(validator structure.Validator) {
	validator.Int("absorptionDuration", n.AbsorptionDuration).InRange(AbsorptionDurationMinimum, AbsorptionDurationMaximum)
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
