package food

import (
	"github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewMeal() *food.Meal {
	datum := food.NewMeal()
	datum.Meal = pointer.FromString(test.RandomStringFromArray(food.MealSize()))
	datum.Snack = pointer.FromString(test.RandomStringFromArray(food.IsASnack()))
	datum.Fat = pointer.FromString(test.RandomStringFromArray(food.IsFat()))
	return datum
}

func CloneMeal(datum *food.Meal) *food.Meal {
	if datum == nil {
		return nil
	}
	clone := food.NewMeal()
	clone.Meal = pointer.CloneString(datum.Meal)
	clone.Snack = pointer.CloneString(datum.Snack)
	clone.Fat = pointer.CloneString(datum.Fat)
	return clone
}
