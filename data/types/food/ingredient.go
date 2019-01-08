package food

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	IngredientArrayLengthMaximum = 100
	IngredientBrandLengthMaximum = 100
	IngredientCodeLengthMaximum  = 100
	IngredientNameLengthMaximum  = 100
)

type Ingredient struct {
	Amount      *Amount          `json:"amount,omitempty" bson:"amount,omitempty"`
	Brand       *string          `json:"brand,omitempty" bson:"brand,omitempty"`
	Code        *string          `json:"code,omitempty" bson:"code,omitempty"`
	Ingredients *IngredientArray `json:"ingredients,omitempty" bson:"ingredients,omitempty"`
	Name        *string          `json:"name,omitempty" bson:"name,omitempty"`
	Nutrition   *Nutrition       `json:"nutrition,omitempty" bson:"nutrition,omitempty"`
}

func ParseIngredient(parser structure.ObjectParser) *Ingredient {
	if !parser.Exists() {
		return nil
	}
	datum := NewIngredient()
	parser.Parse(datum)
	return datum
}

func NewIngredient() *Ingredient {
	return &Ingredient{}
}

func (i *Ingredient) Parse(parser structure.ObjectParser) {
	i.Amount = ParseAmount(parser.WithReferenceObjectParser("amount"))
	i.Brand = parser.String("brand")
	i.Code = parser.String("code")
	i.Ingredients = ParseIngredientArray(parser.WithReferenceArrayParser("ingredients"))
	i.Name = parser.String("name")
	i.Nutrition = ParseNutrition(parser.WithReferenceObjectParser("nutrition"))
}

func (i *Ingredient) Validate(validator structure.Validator) {
	if i.Amount != nil {
		i.Amount.Validate(validator.WithReference("amount"))
	}
	validator.String("brand", i.Brand).NotEmpty().LengthLessThanOrEqualTo(IngredientBrandLengthMaximum)
	validator.String("code", i.Code).NotEmpty().LengthLessThanOrEqualTo(IngredientCodeLengthMaximum)
	if i.Ingredients != nil {
		i.Ingredients.Validate(validator.WithReference("ingredients"))
	}
	validator.String("name", i.Name).NotEmpty().LengthLessThanOrEqualTo(IngredientNameLengthMaximum)
	if i.Nutrition != nil {
		i.Nutrition.Validate(validator.WithReference("nutrition"))
	}
}

func (i *Ingredient) Normalize(normalizer data.Normalizer) {
	if i.Amount != nil {
		i.Amount.Normalize(normalizer.WithReference("amount"))
	}
	if i.Ingredients != nil {
		i.Ingredients.Normalize(normalizer.WithReference("ingredients"))
	}
	if i.Nutrition != nil {
		i.Nutrition.Normalize(normalizer.WithReference("nutrition"))
	}
}

type IngredientArray []*Ingredient

func ParseIngredientArray(parser structure.ArrayParser) *IngredientArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewIngredientArray()
	parser.Parse(datum)
	return datum
}

func NewIngredientArray() *IngredientArray {
	return &IngredientArray{}
}

func (i *IngredientArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*i = append(*i, ParseIngredient(parser.WithReferenceObjectParser(reference)))
	}
}

func (i *IngredientArray) Validate(validator structure.Validator) {
	if length := len(*i); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	} else if length > IngredientArrayLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, IngredientArrayLengthMaximum))
	}
	for index, datum := range *i {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (i *IngredientArray) Normalize(normalizer data.Normalizer) {
	for index, datum := range *i {
		if datum != nil {
			datum.Normalize(normalizer.WithReference(strconv.Itoa(index)))
		}
	}
}
