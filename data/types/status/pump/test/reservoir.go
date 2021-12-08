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

func NewObjectFromReservoir(datum *dataTypesStatusPump.Reservoir, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Time != nil {
		object["time"] = test.NewObjectFromTime(*datum.Time, objectFormat)
	}
	if datum.Remaining != nil {
		object["remaining"] = test.NewObjectFromFloat64(*datum.Remaining, objectFormat)
	}
	if datum.Units != nil {
		object["units"] = test.NewObjectFromString(*datum.Units, objectFormat)
	}
	return object
}
