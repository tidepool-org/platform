package test

import (
	dataTypesDosingdecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomBolus() *dataTypesDosingdecision.Bolus {
	datum := dataTypesDosingdecision.NewBolus()
	if test.RandomBool() {
		datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesDosingdecision.BolusDurationMinimum, dataTypesDosingdecision.BolusDurationMaximum))
		datum.Extended = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingdecision.BolusExtendedMinimum, dataTypesDosingdecision.BolusExtendedMaximum))
	}
	if datum.Extended == nil || test.RandomBool() {
		datum.Normal = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingdecision.BolusNormalMinimum, dataTypesDosingdecision.BolusNormalMaximum))
	}
	return datum
}

func RandomBolusDEPRECATED() *dataTypesDosingdecision.Bolus {
	datum := dataTypesDosingdecision.NewBolus()
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingdecision.BolusAmountMinimum, dataTypesDosingdecision.BolusAmountMaximum))
	return datum
}

func CloneBolus(datum *dataTypesDosingdecision.Bolus) *dataTypesDosingdecision.Bolus {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingdecision.NewBolus()
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	clone.Duration = pointer.CloneInt(datum.Duration)
	clone.Extended = pointer.CloneFloat64(datum.Extended)
	clone.Normal = pointer.CloneFloat64(datum.Normal)
	return clone
}

func NewObjectFromBolus(datum *dataTypesDosingdecision.Bolus, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.Amount != nil {
		object["amount"] = test.NewObjectFromFloat64(*datum.Amount, objectFormat)
	}
	if datum.Duration != nil {
		object["duration"] = test.NewObjectFromInt(*datum.Duration, objectFormat)
	}
	if datum.Extended != nil {
		object["extended"] = test.NewObjectFromFloat64(*datum.Extended, objectFormat)
	}
	if datum.Normal != nil {
		object["normal"] = test.NewObjectFromFloat64(*datum.Normal, objectFormat)
	}
	return object
}
