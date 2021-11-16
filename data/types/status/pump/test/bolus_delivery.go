package test

import (
	dataTypesStatusPump "github.com/tidepool-org/platform/data/types/status/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomBolusDelivery() *dataTypesStatusPump.BolusDelivery {
	state := test.RandomStringFromArray(dataTypesStatusPump.BolusDeliveryStates())
	datum := dataTypesStatusPump.NewBolusDelivery()
	datum.State = pointer.FromString(state)
	switch state {
	case dataTypesStatusPump.BolusDeliveryStateDelivering:
		datum.Dose = RandomBolusDose()
	}
	return datum
}

func CloneBolusDelivery(datum *dataTypesStatusPump.BolusDelivery) *dataTypesStatusPump.BolusDelivery {
	if datum == nil {
		return nil
	}
	clone := dataTypesStatusPump.NewBolusDelivery()
	clone.State = pointer.CloneString(datum.State)
	clone.Dose = CloneBolusDose(datum.Dose)
	return clone
}

func RandomBolusDose() *dataTypesStatusPump.BolusDose {
	datum := dataTypesStatusPump.NewBolusDose()
	datum.StartTime = pointer.FromTime(test.RandomTime())
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesStatusPump.BolusDoseAmountMinimum, dataTypesStatusPump.BolusDoseAmountMaximum))
	datum.AmountDelivered = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesStatusPump.BolusDoseAmountDeliveredMinimum, dataTypesStatusPump.BolusDoseAmountDeliveredMaximum))
	return datum
}

func CloneBolusDose(datum *dataTypesStatusPump.BolusDose) *dataTypesStatusPump.BolusDose {
	if datum == nil {
		return nil
	}
	clone := dataTypesStatusPump.NewBolusDose()
	clone.StartTime = pointer.CloneTime(datum.StartTime)
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	clone.AmountDelivered = pointer.CloneFloat64(datum.AmountDelivered)
	return clone
}
