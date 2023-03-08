package food

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	SmallMeal    = "small"
	MediumMeal   = "medium"
	LargeMeal    = "large"
	Snack        = "yes"
	NotASnack    = "no"
	FatMeal      = "yes"
	NotAFatMeal  = "no"
	UmmSource    = "umm"
	ManualSource = "manual"
)

func MealSize() []string {
	return []string{
		SmallMeal,
		MediumMeal,
		LargeMeal,
	}
}

func IsASnack() []string {
	return []string{
		Snack,
		NotASnack,
	}
}

func IsFat() []string {
	return []string{
		FatMeal,
		NotAFatMeal,
	}
}

func MealSource() []string {
	return []string{
		UmmSource,
		ManualSource,
	}
}

type Meal struct {
	Meal   *string `json:"meal,omitempty" bson:"meal,omitempty"`
	Snack  *string `json:"snack,omitempty" bson:"snack,omitempty"`
	Fat    *string `json:"fat,omitempty" bson:"fat,omitempty"`
	Source *string `json:"source,omitempty" bson:"source,omitempty"`
}

func ParseMeal(parser structure.ObjectParser) *Meal {
	if !parser.Exists() {
		return nil
	}
	datum := NewMeal()
	parser.Parse(datum)
	return datum
}

func NewMeal() *Meal {
	return &Meal{}
}

func (m *Meal) Parse(parser structure.ObjectParser) {
	m.Meal = parser.String("meal")
	m.Snack = parser.String("snack")
	m.Fat = parser.String("fat")
	m.Source = parser.String("source")
}

func (m *Meal) Validate(validator structure.Validator) {
	validator.String("meal", m.Meal).OneOf(MealSize()...)
	validator.String("snack", m.Snack).OneOf(IsASnack()...)
	validator.String("fat", m.Fat).OneOf(IsFat()...)
	validator.String("source", m.Source).OneOf(MealSource()...)
}

func (m *Meal) Normalize(normalizer data.Normalizer) {}
