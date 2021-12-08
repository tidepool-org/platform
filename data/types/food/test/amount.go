package test

import (
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomAmount() *dataTypesFood.Amount {
	datum := dataTypesFood.NewAmount()
	datum.Units = pointer.FromString(test.RandomStringFromRange(1, dataTypesFood.AmountUnitsLengthMaximum))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesFood.AmountValueMinimum, test.RandomFloat64Maximum()))
	return datum
}

func CloneAmount(datum *dataTypesFood.Amount) *dataTypesFood.Amount {
	if datum == nil {
		return nil
	}
	clone := dataTypesFood.NewAmount()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

func NewObjectFromAmount(datum *dataTypesFood.Amount, objectFormat test.ObjectFormat) map[string]interface{} {
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
