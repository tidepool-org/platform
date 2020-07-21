package test

import (
	dataTypesPumpStatus "github.com/tidepool-org/platform/data/types/pumpstatus"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomBolusDelivery() *dataTypesPumpStatus.BolusDelivery {
	state := test.RandomStringFromArray(dataTypesPumpStatus.BolusDeliveryStates())
	datum := dataTypesPumpStatus.NewBolusDelivery()
	datum.State = pointer.FromString(state)
	switch state {
	case dataTypesPumpStatus.BolusDeliveryStateDelivering:
		datum.Dose = RandomBolusDose()
	}
	return datum
}

func CloneBolusDelivery(datum *dataTypesPumpStatus.BolusDelivery) *dataTypesPumpStatus.BolusDelivery {
	if datum == nil {
		return nil
	}
	clone := dataTypesPumpStatus.NewBolusDelivery()
	clone.State = pointer.CloneString(datum.State)
	clone.Dose = CloneBolusDose(datum.Dose)
	return clone
}

func RandomBolusDose() *dataTypesPumpStatus.BolusDose {
	datum := dataTypesPumpStatus.NewBolusDose()
	datum.StartTime = pointer.FromTime(test.RandomTime())
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesPumpStatus.BolusDoseAmountMinimum, dataTypesPumpStatus.BolusDoseAmountMaximum))
	datum.AmountDelivered = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesPumpStatus.BolusDoseAmountDeliveredMinimum, dataTypesPumpStatus.BolusDoseAmountDeliveredMaximum))
	return datum
}

func CloneBolusDose(datum *dataTypesPumpStatus.BolusDose) *dataTypesPumpStatus.BolusDose {
	if datum == nil {
		return nil
	}
	clone := dataTypesPumpStatus.NewBolusDose()
	clone.StartTime = pointer.CloneTime(datum.StartTime)
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	clone.AmountDelivered = pointer.CloneFloat64(datum.AmountDelivered)
	return clone
}
