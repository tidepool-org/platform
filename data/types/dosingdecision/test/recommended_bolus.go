package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomRecommendedBolus() *dataTypesDosingDecision.RecommendedBolus {
	datum := dataTypesDosingDecision.NewRecommendedBolus()
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingDecision.RecommendedBolusAmountMinimum, dataTypesDosingDecision.RecommendedBolusAmountMaximum))
	return datum
}

func CloneRecommendedBolus(datum *dataTypesDosingDecision.RecommendedBolus) *dataTypesDosingDecision.RecommendedBolus {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewRecommendedBolus()
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	return clone
}
