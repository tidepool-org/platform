package test

import (
	dataTypesStatusPump "github.com/tidepool-org/platform/data/types/status/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomBattery() *dataTypesStatusPump.Battery {
	units := test.RandomStringFromArray(dataTypesStatusPump.BatteryUnits())
	datum := dataTypesStatusPump.NewBattery()
	datum.Time = pointer.FromTime(test.RandomTime())
	switch units {
	case dataTypesStatusPump.BatteryUnitsPercent:
		datum.Remaining = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesStatusPump.BatteryRemainingPercentMinimum, dataTypesStatusPump.BatteryRemainingPercentMaximum))
	}
	datum.Units = pointer.FromString(units)
	return datum
}

func CloneBattery(datum *dataTypesStatusPump.Battery) *dataTypesStatusPump.Battery {
	if datum == nil {
		return nil
	}
	clone := dataTypesStatusPump.NewBattery()
	clone.Time = pointer.CloneTime(datum.Time)
	clone.Remaining = pointer.CloneFloat64(datum.Remaining)
	clone.Units = pointer.CloneString(datum.Units)
	return clone
}
