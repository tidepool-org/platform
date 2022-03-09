package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesFoodTest "github.com/tidepool-org/platform/data/types/food/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomFood() *dataTypesDosingDecision.Food {
	datum := dataTypesDosingDecision.NewFood()
	datum.Time = pointer.FromTime(test.RandomTime())
	datum.Nutrition = dataTypesFoodTest.RandomNutrition()
	return datum
}

func CloneFood(datum *dataTypesDosingDecision.Food) *dataTypesDosingDecision.Food {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewFood()
	clone.Time = pointer.CloneTime(datum.Time)
	clone.Nutrition = dataTypesFoodTest.CloneNutrition(datum.Nutrition)
	return clone
}

func NewObjectFromFood(datum *dataTypesDosingDecision.Food, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Time != nil {
		object["time"] = test.NewObjectFromTime(*datum.Time, objectFormat)
	}
	if datum.Nutrition != nil {
		object["nutrition"] = dataTypesFoodTest.NewObjectFromNutrition(datum.Nutrition, objectFormat)
	}
	return object
}
