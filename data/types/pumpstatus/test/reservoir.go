package test

import (
	dataTypesPumpStatus "github.com/tidepool-org/platform/data/types/pumpstatus"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomReservoir() *dataTypesPumpStatus.Reservoir {
	units := test.RandomStringFromArray(dataTypesPumpStatus.ReservoirUnits())
	datum := dataTypesPumpStatus.NewReservoir()
	datum.Time = pointer.FromTime(test.RandomTime())
	switch units {
	case dataTypesPumpStatus.ReservoirUnitsUnits:
		datum.Remaining = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesPumpStatus.ReservoirRemainingUnitsMinimum, dataTypesPumpStatus.ReservoirRemainingUnitsMaximum))
	}
	datum.Units = pointer.FromString(units)
	return datum
}

func CloneReservoir(datum *dataTypesPumpStatus.Reservoir) *dataTypesPumpStatus.Reservoir {
	if datum == nil {
		return nil
	}
	clone := dataTypesPumpStatus.NewReservoir()
	clone.Time = pointer.CloneTime(datum.Time)
	clone.Remaining = pointer.CloneFloat64(datum.Remaining)
	clone.Units = pointer.CloneString(datum.Units)
	return clone
}
