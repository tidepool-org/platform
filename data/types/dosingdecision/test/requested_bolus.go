package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomRequestedBolus() *dataTypesDosingDecision.RequestedBolus {
	datum := dataTypesDosingDecision.NewRequestedBolus()
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingDecision.RequestedBolusAmountMinimum, dataTypesDosingDecision.RequestedBolusAmountMaximum))
	return datum
}

func CloneRequestedBolus(datum *dataTypesDosingDecision.RequestedBolus) *dataTypesDosingDecision.RequestedBolus {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewRequestedBolus()
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	return clone
}

func NewObjectFromRequestedBolus(datum *dataTypesDosingDecision.RequestedBolus, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Amount != nil {
		object["amount"] = test.NewObjectFromFloat64(*datum.Amount, objectFormat)
	}
	return object
}
