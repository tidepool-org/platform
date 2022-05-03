package test

import (
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomCarbohydrate() *dataTypesFood.Carbohydrate {
	units := test.RandomStringFromArray(dataTypesFood.CarbohydrateUnits())
	datum := dataTypesFood.NewCarbohydrate()
	switch units {
	case dataTypesFood.CarbohydrateUnitsGrams:
		datum.DietaryFiber = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.CarbohydrateDietaryFiberGramsMinimum, dataTypesFood.CarbohydrateDietaryFiberGramsMaximum))
		datum.Net = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.CarbohydrateNetGramsMinimum, dataTypesFood.CarbohydrateNetGramsMaximum))
		datum.Sugars = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.CarbohydrateSugarsGramsMinimum, dataTypesFood.CarbohydrateSugarsGramsMaximum))
		datum.Total = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.CarbohydrateTotalGramsMinimum, dataTypesFood.CarbohydrateTotalGramsMaximum))
	}
	datum.Units = pointer.FromString(units)
	return datum
}

func CloneCarbohydrate(datum *dataTypesFood.Carbohydrate) *dataTypesFood.Carbohydrate {
	if datum == nil {
		return nil
	}
	clone := dataTypesFood.NewCarbohydrate()
	clone.DietaryFiber = pointer.CloneFloat64(datum.DietaryFiber)
	clone.Net = pointer.CloneFloat64(datum.Net)
	clone.Sugars = pointer.CloneFloat64(datum.Sugars)
	clone.Total = pointer.CloneFloat64(datum.Total)
	clone.Units = pointer.CloneString(datum.Units)
	return clone
}

func NewObjectFromCarbohydrate(datum *dataTypesFood.Carbohydrate, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.DietaryFiber != nil {
		object["dietaryFiber"] = test.NewObjectFromFloat64(*datum.DietaryFiber, objectFormat)
	}
	if datum.Net != nil {
		object["net"] = test.NewObjectFromFloat64(*datum.Net, objectFormat)
	}
	if datum.Sugars != nil {
		object["sugars"] = test.NewObjectFromFloat64(*datum.Sugars, objectFormat)
	}
	if datum.Total != nil {
		object["total"] = test.NewObjectFromFloat64(*datum.Total, objectFormat)
	}
	if datum.Units != nil {
		object["units"] = test.NewObjectFromString(*datum.Units, objectFormat)
	}
	return object
}
