package test

import (
	dataTypesStatusPump "github.com/tidepool-org/platform/data/types/status/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomBasalDelivery() *dataTypesStatusPump.BasalDelivery {
	state := test.RandomStringFromArray(dataTypesStatusPump.BasalDeliveryStates())
	datum := dataTypesStatusPump.NewBasalDelivery()
	datum.State = pointer.FromString(state)
	switch state {
	case dataTypesStatusPump.BasalDeliveryStateScheduled, dataTypesStatusPump.BasalDeliveryStateSuspended:
		datum.Time = pointer.FromTime(test.RandomTime())
	case dataTypesStatusPump.BasalDeliveryStateTemporary:
		datum.Dose = RandomBasalDose()
	}
	return datum
}

func CloneBasalDelivery(datum *dataTypesStatusPump.BasalDelivery) *dataTypesStatusPump.BasalDelivery {
	if datum == nil {
		return nil
	}
	clone := dataTypesStatusPump.NewBasalDelivery()
	clone.State = pointer.CloneString(datum.State)
	clone.Time = pointer.CloneTime(datum.Time)
	clone.Dose = CloneBasalDose(datum.Dose)
	return clone
}

func NewObjectFromBasalDelivery(datum *dataTypesStatusPump.BasalDelivery, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.State != nil {
		object["state"] = test.NewObjectFromString(*datum.State, objectFormat)
	}
	if datum.Time != nil {
		object["time"] = test.NewObjectFromTime(*datum.Time, objectFormat)
	}
	if datum.Dose != nil {
		object["dose"] = NewObjectFromBasalDose(datum.Dose, objectFormat)
	}
	return object
}

func RandomBasalDose() *dataTypesStatusPump.BasalDose {
	datum := dataTypesStatusPump.NewBasalDose()
	datum.StartTime = pointer.FromTime(test.RandomTime())
	datum.EndTime = pointer.FromTime(test.RandomTimeAfter(*datum.StartTime))
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesStatusPump.BasalDoseRateMinimum, dataTypesStatusPump.BasalDoseRateMaximum))
	datum.AmountDelivered = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesStatusPump.BasalDoseAmountDeliveredMinimum, dataTypesStatusPump.BasalDoseAmountDeliveredMaximum))
	return datum
}

func CloneBasalDose(datum *dataTypesStatusPump.BasalDose) *dataTypesStatusPump.BasalDose {
	if datum == nil {
		return nil
	}
	clone := dataTypesStatusPump.NewBasalDose()
	clone.StartTime = pointer.CloneTime(datum.StartTime)
	clone.EndTime = pointer.CloneTime(datum.EndTime)
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	clone.AmountDelivered = pointer.CloneFloat64(datum.AmountDelivered)
	return clone
}

func NewObjectFromBasalDose(datum *dataTypesStatusPump.BasalDose, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.StartTime != nil {
		object["startTime"] = test.NewObjectFromTime(*datum.StartTime, objectFormat)
	}
	if datum.EndTime != nil {
		object["endTime"] = test.NewObjectFromTime(*datum.EndTime, objectFormat)
	}
	if datum.Rate != nil {
		object["rate"] = test.NewObjectFromFloat64(*datum.Rate, objectFormat)
	}
	if datum.AmountDelivered != nil {
		object["amountDelivered"] = test.NewObjectFromFloat64(*datum.AmountDelivered, objectFormat)
	}
	return object
}
