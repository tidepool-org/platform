package test

import (
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomIngredient(ingredientArrayDepthLimit int) *dataTypesFood.Ingredient {
	datum := dataTypesFood.NewIngredient()
	datum.Amount = RandomAmount()
	datum.Brand = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.IngredientBrandLengthMaximum))
	datum.Code = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.IngredientCodeLengthMaximum))
	datum.Ingredients = RandomIngredientArray(ingredientArrayDepthLimit)
	datum.Name = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.IngredientNameLengthMaximum))
	datum.Nutrition = RandomNutrition()
	return datum
}

func CloneIngredient(datum *dataTypesFood.Ingredient) *dataTypesFood.Ingredient {
	if datum == nil {
		return nil
	}
	clone := dataTypesFood.NewIngredient()
	clone.Amount = CloneAmount(datum.Amount)
	clone.Brand = pointer.CloneString(datum.Brand)
	clone.Code = pointer.CloneString(datum.Code)
	clone.Ingredients = CloneIngredientArray(datum.Ingredients)
	clone.Name = pointer.CloneString(datum.Name)
	clone.Nutrition = CloneNutrition(datum.Nutrition)
	return clone
}

func NewObjectFromIngredient(datum *dataTypesFood.Ingredient, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Amount != nil {
		object["amount"] = NewObjectFromAmount(datum.Amount, objectFormat)
	}
	if datum.Brand != nil {
		object["brand"] = test.NewObjectFromString(*datum.Brand, objectFormat)
	}
	if datum.Code != nil {
		object["code"] = test.NewObjectFromString(*datum.Code, objectFormat)
	}
	if datum.Ingredients != nil {
		object["ingredients"] = NewArrayFromIngredientArray(datum.Ingredients, objectFormat)
	}
	if datum.Name != nil {
		object["name"] = test.NewObjectFromString(*datum.Name, objectFormat)
	}
	if datum.Nutrition != nil {
		object["nutrition"] = NewObjectFromNutrition(datum.Nutrition, objectFormat)
	}
	return object
}

func RandomIngredientArray(ingredientArrayDepthLimit int) *dataTypesFood.IngredientArray {
	if ingredientArrayDepthLimit--; ingredientArrayDepthLimit <= 0 {
		return nil
	}
	datum := dataTypesFood.NewIngredientArray()
	for count := 0; count < test.RandomIntFromRange(1, 3); count++ {
		*datum = append(*datum, RandomIngredient(ingredientArrayDepthLimit))
	}
	return datum
}

func CloneIngredientArray(datumArray *dataTypesFood.IngredientArray) *dataTypesFood.IngredientArray {
	if datumArray == nil {
		return nil
	}
	clone := dataTypesFood.NewIngredientArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneIngredient(datum))
	}
	return clone
}

func NewArrayFromIngredientArray(datumArray *dataTypesFood.IngredientArray, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := []interface{}{}
	for _, datum := range *datumArray {
		array = append(array, NewObjectFromIngredient(datum, objectFormat))
	}
	return array
}

func AnonymizeIngredientArray(datumArray *dataTypesFood.IngredientArray) []interface{} {
	array := []interface{}{}
	for _, datum := range *datumArray {
		array = append(array, datum)
	}
	return array
}
