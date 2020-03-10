package test

import (
	dataTypesPumpStatus "github.com/tidepool-org/platform/data/types/pumpstatus"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomBattery() *dataTypesPumpStatus.Battery {
	units := test.RandomStringFromArray(dataTypesPumpStatus.BatteryUnits())
	datum := dataTypesPumpStatus.NewBattery()
	datum.Time = pointer.FromString(test.RandomTime().Format(dataTypesPumpStatus.TimeFormat))
	switch units {
	case dataTypesPumpStatus.BatteryUnitsPercent:
		datum.Remaining = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesPumpStatus.BatteryRemainingPercentMinimum, dataTypesPumpStatus.BatteryRemainingPercentMaximum))
	}
	datum.Units = pointer.FromString(units)
	return datum
}

func CloneBattery(datum *dataTypesPumpStatus.Battery) *dataTypesPumpStatus.Battery {
	if datum == nil {
		return nil
	}
	clone := dataTypesPumpStatus.NewBattery()
	clone.Time = pointer.CloneString(datum.Time)
	clone.Remaining = pointer.CloneFloat64(datum.Remaining)
	clone.Units = pointer.CloneString(datum.Units)
	return clone
}
