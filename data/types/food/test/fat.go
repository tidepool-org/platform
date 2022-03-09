package test

import (
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomFat() *dataTypesFood.Fat {
	units := test.RandomStringFromArray(dataTypesFood.FatUnits())
	datum := dataTypesFood.NewFat()
	switch units {
	case dataTypesFood.FatUnitsGrams:
		datum.Total = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.FatTotalGramsMinimum, dataTypesFood.FatTotalGramsMaximum))
	}
	datum.Units = pointer.FromString(units)
	return datum
}

func CloneFat(datum *dataTypesFood.Fat) *dataTypesFood.Fat {
	if datum == nil {
		return nil
	}
	clone := dataTypesFood.NewFat()
	clone.Total = pointer.CloneFloat64(datum.Total)
	clone.Units = pointer.CloneString(datum.Units)
	return clone
}

func NewObjectFromFat(datum *dataTypesFood.Fat, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Total != nil {
		object["total"] = test.NewObjectFromFloat64(*datum.Total, objectFormat)
	}
	if datum.Units != nil {
		object["units"] = test.NewObjectFromString(*datum.Units, objectFormat)
	}
	return object
}
