package test

import (
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomProtein() *dataTypesFood.Protein {
	units := test.RandomStringFromArray(dataTypesFood.ProteinUnits())
	datum := dataTypesFood.NewProtein()
	switch units {
	case dataTypesFood.ProteinUnitsGrams:
		datum.Total = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.ProteinTotalGramsMinimum, dataTypesFood.ProteinTotalGramsMaximum))
	}
	datum.Units = pointer.FromString(units)
	return datum
}

func CloneProtein(datum *dataTypesFood.Protein) *dataTypesFood.Protein {
	if datum == nil {
		return nil
	}
	clone := dataTypesFood.NewProtein()
	clone.Total = pointer.CloneFloat64(datum.Total)
	clone.Units = pointer.CloneString(datum.Units)
	return clone
}

func NewObjectFromProtein(datum *dataTypesFood.Protein, objectFormat test.ObjectFormat) map[string]interface{} {
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
