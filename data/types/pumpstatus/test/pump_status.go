package test

import (
	dataTypesPumpStatus "github.com/tidepool-org/platform/data/types/pumpstatus"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
)

func RandomPumpStatus() *dataTypesPumpStatus.PumpStatus {
	datum := dataTypesPumpStatus.New()
	datum.Base = *dataTypesTest.NewBase()
	datum.Type = "pumpStatus"
	datum.BasalDelivery = RandomBasalDelivery()
	datum.Battery = RandomBattery()
	datum.BolusDelivery = RandomBolusDelivery()
	datum.Device = RandomDevice()
	datum.Reservoir = RandomReservoir()
	return datum
}

func ClonePumpStatus(datum *dataTypesPumpStatus.PumpStatus) *dataTypesPumpStatus.PumpStatus {
	if datum == nil {
		return nil
	}
	clone := dataTypesPumpStatus.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.BasalDelivery = CloneBasalDelivery(datum.BasalDelivery)
	clone.Battery = CloneBattery(datum.Battery)
	clone.BolusDelivery = CloneBolusDelivery(datum.BolusDelivery)
	clone.Device = CloneDevice(datum.Device)
	clone.Reservoir = CloneReservoir(datum.Reservoir)
	return clone
}
