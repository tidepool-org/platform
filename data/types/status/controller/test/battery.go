package test

import (
	dataTypesStatusController "github.com/tidepool-org/platform/data/types/status/controller"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomBattery() *dataTypesStatusController.Battery {
	units := test.RandomStringFromArray(dataTypesStatusController.BatteryUnits())
	datum := dataTypesStatusController.NewBattery()
	datum.Time = pointer.FromTime(test.RandomTime())
	datum.State = pointer.FromString(test.RandomStringFromArray(dataTypesStatusController.BatteryStates()))
	switch units {
	case dataTypesStatusController.BatteryUnitsPercent:
		datum.Remaining = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesStatusController.BatteryRemainingPercentMinimum, dataTypesStatusController.BatteryRemainingPercentMaximum))
	}
	datum.Units = pointer.FromString(units)
	return datum
}

func CloneBattery(datum *dataTypesStatusController.Battery) *dataTypesStatusController.Battery {
	if datum == nil {
		return nil
	}
	clone := dataTypesStatusController.NewBattery()
	clone.Time = pointer.CloneTime(datum.Time)
	clone.State = pointer.CloneString(datum.State)
	clone.Remaining = pointer.CloneFloat64(datum.Remaining)
	clone.Units = pointer.CloneString(datum.Units)
	return clone
}

func NewObjectFromBattery(datum *dataTypesStatusController.Battery, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Time != nil {
		object["time"] = test.NewObjectFromTime(*datum.Time, objectFormat)
	}
	if datum.State != nil {
		object["state"] = test.NewObjectFromString(*datum.State, objectFormat)
	}
	if datum.Remaining != nil {
		object["remaining"] = test.NewObjectFromFloat64(*datum.Remaining, objectFormat)
	}
	if datum.Units != nil {
		object["units"] = test.NewObjectFromString(*datum.Units, objectFormat)
	}
	return object
}
