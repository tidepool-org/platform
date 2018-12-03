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

func ParseIngredient(parser data.ObjectParser) *Ingredient {
	if parser.Object() == nil {
		return nil
	}
	datum := NewIngredient()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewIngredient() *Ingredient {
	return &Ingredient{}
}

func (i *Ingredient) Parse(parser data.ObjectParser) {
	i.Amount = ParseAmount(parser.NewChildObjectParser("amount"))
	i.Brand = parser.ParseString("brand")
	i.Code = parser.ParseString("code")
	i.Ingredients = ParseIngredientArray(parser.NewChildArrayParser("ingredients"))
	i.Name = parser.ParseString("name")
	i.Nutrition = ParseNutrition(parser.NewChildObjectParser("nutrition"))
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

func ParseIngredientArray(parser data.ArrayParser) *IngredientArray {
	if parser.Array() == nil {
		return nil
	}
	datum := NewIngredientArray()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewIngredientArray() *IngredientArray {
	return &IngredientArray{}
}

func (i *IngredientArray) Parse(parser data.ArrayParser) {
	for index := range *parser.Array() {
		*i = append(*i, ParseIngredient(parser.NewChildObjectParser(index)))
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
