package test

import (
	dataTypesStatusPump "github.com/tidepool-org/platform/data/types/status/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomReservoir() *dataTypesStatusPump.Reservoir {
	units := test.RandomStringFromArray(dataTypesStatusPump.ReservoirUnits())
	datum := dataTypesStatusPump.NewReservoir()
	datum.Time = pointer.FromTime(test.RandomTime())
	switch units {
	case dataTypesStatusPump.ReservoirUnitsUnits:
		datum.Remaining = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesStatusPump.ReservoirRemainingUnitsMinimum, dataTypesStatusPump.ReservoirRemainingUnitsMaximum))
	}
	datum.Units = pointer.FromString(units)
	return datum
}

func CloneReservoir(datum *dataTypesStatusPump.Reservoir) *dataTypesStatusPump.Reservoir {
	if datum == nil {
		return nil
	}
	clone := dataTypesStatusPump.NewReservoir()
	clone.Time = pointer.CloneTime(datum.Time)
	clone.Remaining = pointer.CloneFloat64(datum.Remaining)
	clone.Units = pointer.CloneString(datum.Units)
	return clone
}
