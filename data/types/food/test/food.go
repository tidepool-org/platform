package test

import (
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomFood(ingredientArrayDepthLimit int) *dataTypesFood.Food {
	datum := randomFood(ingredientArrayDepthLimit)
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "food"
	return datum
}

func RandomFoodForParser(ingredientArrayDepthLimit int) *dataTypesFood.Food {
	datum := randomFood(ingredientArrayDepthLimit)
	datum.Base = *dataTypesTest.RandomBaseForParser()
	datum.Type = "food"
	return datum
}

func randomFood(ingredientArrayDepthLimit int) *dataTypesFood.Food {
	datum := dataTypesFood.New()
	datum.Amount = RandomAmount()
	datum.Brand = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.BrandLengthMaximum))
	datum.Code = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.CodeLengthMaximum))
	datum.Ingredients = RandomIngredientArray(ingredientArrayDepthLimit)
	datum.Meal = pointer.FromString(test.RandomStringFromArray(dataTypesFood.Meals()))
	if datum.Meal != nil && *datum.Meal == dataTypesFood.MealOther {
		datum.MealOther = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.MealOtherLengthMaximum))
	}
	datum.Name = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.NameLengthMaximum))
	datum.Nutrition = RandomNutrition()
	return datum
}

func CloneFood(datum *dataTypesFood.Food) *dataTypesFood.Food {
	if datum == nil {
		return nil
	}
	clone := dataTypesFood.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.Amount = CloneAmount(datum.Amount)
	clone.Brand = pointer.CloneString(datum.Brand)
	clone.Code = pointer.CloneString(datum.Code)
	clone.Ingredients = CloneIngredientArray(datum.Ingredients)
	clone.Meal = pointer.CloneString(datum.Meal)
	clone.MealOther = pointer.CloneString(datum.MealOther)
	clone.Name = pointer.CloneString(datum.Name)
	clone.Nutrition = CloneNutrition(datum.Nutrition)
	return clone
}

func NewObjectFromFood(datum *dataTypesFood.Food, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := dataTypesTest.NewObjectFromBase(&datum.Base, objectFormat)
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
	if datum.Meal != nil {
		object["meal"] = test.NewObjectFromString(*datum.Meal, objectFormat)
	}
	if datum.MealOther != nil {
		object["mealOther"] = test.NewObjectFromString(*datum.MealOther, objectFormat)
	}
	if datum.Name != nil {
		object["name"] = test.NewObjectFromString(*datum.Name, objectFormat)
	}
	if datum.Nutrition != nil {
		object["nutrition"] = NewObjectFromNutrition(datum.Nutrition, objectFormat)
	}
	return object
}
