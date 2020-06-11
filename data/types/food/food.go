package food

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "food"

	BrandLengthMaximum     = 100
	CodeLengthMaximum      = 100
	MealBreakfast          = "breakfast"
	MealDinner             = "dinner"
	MealLunch              = "lunch"
	MealOther              = "other"
	MealOtherLengthMaximum = 100
	MealSnack              = "snack"
	MealRescueCarbs        = "rescuecarbs"
	NameLengthMaximum      = 100
)

func Meals() []string {
	return []string{
		MealBreakfast,
		MealDinner,
		MealLunch,
		MealOther,
		MealSnack,
		MealRescueCarbs,
	}
}

type Food struct {
	types.Base `bson:",inline"`

	Amount              *Amount             `json:"amount,omitempty" bson:"amount,omitempty"`
	Brand               *string             `json:"brand,omitempty" bson:"brand,omitempty"`
	Code                *string             `json:"code,omitempty" bson:"code,omitempty"`
	Ingredients         *IngredientArray    `json:"ingredients,omitempty" bson:"ingredients,omitempty"`
	Meal                *string             `json:"meal,omitempty" bson:"meal,omitempty"`
	MealOther           *string             `json:"mealOther,omitempty" bson:"mealOther,omitempty"`
	Name                *string             `json:"name,omitempty" bson:"name,omitempty"`
	Nutrition           *Nutrition          `json:"nutrition,omitempty" bson:"nutrition,omitempty"`
	PrescribedNutrition *Nutrition          `json:"prescribedNutrition,omitempty" bson:"prescribedNutrition,omitempty"`
	Prescriptor         *common.Prescriptor `bson:",inline"`
}

func New() *Food {
	return &Food{
		Base:        types.New(Type),
		Prescriptor: common.NewPrescriptor(),
	}
}

func (f *Food) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(f.Meta())
	}

	f.Base.Parse(parser)

	f.Amount = ParseAmount(parser.WithReferenceObjectParser("amount"))
	f.Brand = parser.String("brand")
	f.Code = parser.String("code")
	f.Ingredients = ParseIngredientArray(parser.WithReferenceArrayParser("ingredients"))
	f.Meal = parser.String("meal")
	f.MealOther = parser.String("mealOther")
	f.Name = parser.String("name")
	f.Nutrition = ParseNutrition(parser.WithReferenceObjectParser("nutrition"))
	f.PrescribedNutrition = ParseNutrition(parser.WithReferenceObjectParser("prescribedNutrition"))
	f.Prescriptor.Parse(parser)
}

func (f *Food) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(f.Meta())
	}

	f.Base.Validate(validator)

	if f.Type != "" {
		validator.String("type", &f.Type).EqualTo(Type)
	}

	if f.Amount != nil {
		f.Amount.Validate(validator.WithReference("amount"))
	}
	validator.String("brand", f.Brand).NotEmpty().LengthLessThanOrEqualTo(BrandLengthMaximum)
	validator.String("code", f.Code).NotEmpty().LengthLessThanOrEqualTo(CodeLengthMaximum)
	if f.Ingredients != nil {
		f.Ingredients.Validate(validator.WithReference("ingredients"))
	}
	validator.String("meal", f.Meal).OneOf(Meals()...)
	if f.Meal != nil && *f.Meal == MealOther {
		validator.String("mealOther", f.MealOther).Exists().NotEmpty().LengthLessThanOrEqualTo(MealOtherLengthMaximum)
	} else {
		validator.String("mealOther", f.MealOther).NotExists()
	}
	validator.String("name", f.Name).NotEmpty().LengthLessThanOrEqualTo(NameLengthMaximum)
	if f.Nutrition != nil {
		f.Nutrition.Validate(validator.WithReference("nutrition"))
	}
	if f.Meal != nil && *f.Meal == MealRescueCarbs {
		if f.Prescriptor != nil {
			f.Prescriptor.Validate(validator)
			if *f.Prescriptor.Prescriptor == common.HybridPrescriptor && f.PrescribedNutrition == nil {
				// Prescribed Nutrition is mandatory
				validator.WithReference("prescribedNutrition").ReportError(structureValidator.ErrorValueNotExists())
			}
		}
		if f.PrescribedNutrition != nil {
			f.PrescribedNutrition.Validate(validator.WithReference("nutrition"))
		}
	}
}

// IsValid returns true if there is no error in the validator
func (f *Food) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (f *Food) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(f.Meta())
	}

	f.Base.Normalize(normalizer)

	if f.Meal != nil && *f.Meal != MealRescueCarbs {
		// Prescriptors are ignored for other types of meals
		f.Prescriptor = nil
		f.PrescribedNutrition = nil
	}

	if f.Amount != nil {
		f.Amount.Normalize(normalizer.WithReference("amount"))
	}
	if f.Ingredients != nil {
		f.Ingredients.Normalize(normalizer.WithReference("ingredients"))
	}
	if f.Nutrition != nil {
		f.Nutrition.Normalize(normalizer.WithReference("nutrition"))
	}
	if f.Prescriptor != nil && *f.Prescriptor.Prescriptor == common.HybridPrescriptor && f.PrescribedNutrition != nil {
		f.PrescribedNutrition.Normalize(normalizer.WithReference("nutrition"))
	} else {
		f.PrescribedNutrition = nil
	}
}
