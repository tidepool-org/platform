package test

import (
	dataTypesStatusPump "github.com/tidepool-org/platform/data/types/status/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
)

func RandomPump() *dataTypesStatusPump.Pump {
	datum := dataTypesStatusPump.New()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "pumpStatus"
	datum.BasalDelivery = RandomBasalDelivery()
	datum.Battery = RandomBattery()
	datum.BolusDelivery = RandomBolusDelivery()
	datum.Device = RandomDevice()
	datum.Reservoir = RandomReservoir()
	return datum
}

func ClonePump(datum *dataTypesStatusPump.Pump) *dataTypesStatusPump.Pump {
	if datum == nil {
		return nil
	}
	clone := dataTypesStatusPump.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.BasalDelivery = CloneBasalDelivery(datum.BasalDelivery)
	clone.Battery = CloneBattery(datum.Battery)
	clone.BolusDelivery = CloneBolusDelivery(datum.BolusDelivery)
	clone.Device = CloneDevice(datum.Device)
	clone.Reservoir = CloneReservoir(datum.Reservoir)
	return clone
}
