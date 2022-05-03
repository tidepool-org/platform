package test

import (
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomNutrition() *dataTypesFood.Nutrition {
	datum := dataTypesFood.NewNutrition()
	datum.EstimatedAbsorptionDuration = pointer.FromInt(test.RandomIntFromRange(dataTypesFood.EstimatedAbsorptionDurationSecondsMinimum, dataTypesFood.EstimatedAbsorptionDurationSecondsMaximum))
	datum.Carbohydrate = RandomCarbohydrate()
	datum.Energy = RandomEnergy()
	datum.Fat = RandomFat()
	datum.Protein = RandomProtein()
	return datum
}

func CloneNutrition(datum *dataTypesFood.Nutrition) *dataTypesFood.Nutrition {
	if datum == nil {
		return nil
	}
	clone := dataTypesFood.NewNutrition()
	clone.EstimatedAbsorptionDuration = pointer.CloneInt(datum.EstimatedAbsorptionDuration)
	clone.Carbohydrate = CloneCarbohydrate(datum.Carbohydrate)
	clone.Energy = CloneEnergy(datum.Energy)
	clone.Fat = CloneFat(datum.Fat)
	clone.Protein = CloneProtein(datum.Protein)
	return clone
}

func NewObjectFromNutrition(datum *dataTypesFood.Nutrition, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.EstimatedAbsorptionDuration != nil {
		object["estimatedAbsorptionDuration"] = test.NewObjectFromInt(*datum.EstimatedAbsorptionDuration, objectFormat)
	}
	if datum.Carbohydrate != nil {
		object["carbohydrate"] = NewObjectFromCarbohydrate(datum.Carbohydrate, objectFormat)
	}
	if datum.Energy != nil {
		object["energy"] = NewObjectFromEnergy(datum.Energy, objectFormat)
	}
	if datum.Fat != nil {
		object["fat"] = NewObjectFromFat(datum.Fat, objectFormat)
	}
	if datum.Protein != nil {
		object["protein"] = NewObjectFromProtein(datum.Protein, objectFormat)
	}
	return object
}
