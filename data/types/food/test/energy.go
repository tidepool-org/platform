package test

import (
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomEnergy() *dataTypesFood.Energy {
	units := test.RandomStringFromArray(dataTypesFood.EnergyUnits())
	datum := dataTypesFood.NewEnergy()
	datum.Units = pointer.FromString(units)
	switch units {
	case dataTypesFood.EnergyUnitsCalories:
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.EnergyValueCaloriesMinimum, dataTypesFood.EnergyValueCaloriesMaximum))
	case dataTypesFood.EnergyUnitsJoules:
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.EnergyValueJoulesMinimum, dataTypesFood.EnergyValueJoulesMaximum))
	case dataTypesFood.EnergyUnitsKilocalories:
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.EnergyValueKilocaloriesMinimum, dataTypesFood.EnergyValueKilocaloriesMaximum))
	case dataTypesFood.EnergyUnitsKilojoules:
		datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.EnergyValueKilojoulesMinimum, dataTypesFood.EnergyValueKilojoulesMaximum))
	}
	return datum
}

func CloneEnergy(datum *dataTypesFood.Energy) *dataTypesFood.Energy {
	if datum == nil {
		return nil
	}
	clone := dataTypesFood.NewEnergy()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

func NewObjectFromEnergy(datum *dataTypesFood.Energy, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Units != nil {
		object["units"] = test.NewObjectFromString(*datum.Units, objectFormat)
	}
	if datum.Value != nil {
		object["value"] = test.NewObjectFromFloat64(*datum.Value, objectFormat)
	}
	return object
}
