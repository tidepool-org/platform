package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomRecommendedBasal() *dataTypesDosingDecision.RecommendedBasal {
	datum := dataTypesDosingDecision.NewRecommendedBasal()
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingDecision.RecommendedBasalRateMinimum, dataTypesDosingDecision.RecommendedBasalRateMaximum))
	datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesDosingDecision.RecommendedBasalDurationMinimum, dataTypesDosingDecision.RecommendedBasalDurationMaximum))
	return datum
}

func CloneRecommendedBasal(datum *dataTypesDosingDecision.RecommendedBasal) *dataTypesDosingDecision.RecommendedBasal {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewRecommendedBasal()
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	clone.Duration = pointer.CloneInt(datum.Duration)
	return clone
}
