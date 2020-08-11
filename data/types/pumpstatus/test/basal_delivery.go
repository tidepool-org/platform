package test

import (
	dataTypesPumpStatus "github.com/tidepool-org/platform/data/types/pumpstatus"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomBasalDelivery() *dataTypesPumpStatus.BasalDelivery {
	state := test.RandomStringFromArray(dataTypesPumpStatus.BasalDeliveryStates())
	datum := dataTypesPumpStatus.NewBasalDelivery()
	datum.State = pointer.FromString(state)
	switch state {
	case dataTypesPumpStatus.BasalDeliveryStateScheduled, dataTypesPumpStatus.BasalDeliveryStateSuspended:
		datum.Time = pointer.FromTime(test.RandomTime())
	case dataTypesPumpStatus.BasalDeliveryStateTemporary:
		datum.Dose = RandomBasalDose()
	}
	return datum
}

func CloneBasalDelivery(datum *dataTypesPumpStatus.BasalDelivery) *dataTypesPumpStatus.BasalDelivery {
	if datum == nil {
		return nil
	}
	clone := dataTypesPumpStatus.NewBasalDelivery()
	clone.State = pointer.CloneString(datum.State)
	clone.Time = pointer.CloneTime(datum.Time)
	clone.Dose = CloneBasalDose(datum.Dose)
	return clone
}

func RandomBasalDose() *dataTypesPumpStatus.BasalDose {
	datum := dataTypesPumpStatus.NewBasalDose()
	datum.StartTime = pointer.FromTime(test.RandomTime())
	datum.EndTime = pointer.FromTime(test.RandomTimeFromRange(*datum.StartTime, test.RandomTimeMaximum()))
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesPumpStatus.BasalDoseRateMinimum, dataTypesPumpStatus.BasalDoseRateMaximum))
	datum.AmountDelivered = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesPumpStatus.BasalDoseAmountDeliveredMinimum, dataTypesPumpStatus.BasalDoseAmountDeliveredMaximum))
	return datum
}

func CloneBasalDose(datum *dataTypesPumpStatus.BasalDose) *dataTypesPumpStatus.BasalDose {
	if datum == nil {
		return nil
	}
	clone := dataTypesPumpStatus.NewBasalDose()
	clone.StartTime = pointer.CloneTime(datum.StartTime)
	clone.EndTime = pointer.CloneTime(datum.EndTime)
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	clone.AmountDelivered = pointer.CloneFloat64(datum.AmountDelivered)
	return clone
}
