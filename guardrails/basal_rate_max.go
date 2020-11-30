package guardrails

import (
	devices "github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
)

const maxBasalRateMaximumDividend = 70

func ValidateBasalRateMaximum(basalRateMaximum pump.BasalRateMaximum, basalRateSchedule *pump.BasalRateStartArray, carbohydrateRatioSchedule *pump.CarbohydrateRatioStartArray, guardRail *devices.BasalRateMaximumGuardRail, validator structure.Validator) {
	validValues := generateBasalRateMaximumValidValues(carbohydrateRatioSchedule, basalRateSchedule, guardRail)
	ValidateValueIfNotNil(basalRateMaximum.Value, validValues, validator.WithReference("value"))
}

func generateBasalRateMaximumValidValues(carbohydrateRatioSchedule *pump.CarbohydrateRatioStartArray, basalRateSchedule *pump.BasalRateStartArray, guardRail *devices.BasalRateMaximumGuardRail) []float64 {
	validValues := generateValidValuesFromAbsoluteBounds(guardRail.AbsoluteBounds)
	if lowestScheduledCarbRatio := getLowestScheduledCarbRatio(carbohydrateRatioSchedule); lowestScheduledCarbRatio != nil {
		max := maxBasalRateMaximumDividend / *lowestScheduledCarbRatio
		validValues = discardValuesLargerThan(validValues, max)
	}
	if highestScheduledBasalRate := getHighestScheduledBasalRate(basalRateSchedule); highestScheduledBasalRate != nil {
		validValues = discardValuesSmallerThan(validValues, *highestScheduledBasalRate)
	}
	return validValues
}

func getLowestScheduledCarbRatio(carbohydrateRatioSchedule *pump.CarbohydrateRatioStartArray) *float64 {
	if carbohydrateRatioSchedule == nil {
		return nil
	}
	var min *float64
	for _, c := range *carbohydrateRatioSchedule {
		if c != nil && c.Amount != nil && (min == nil || *c.Amount < *min) {
			min = c.Amount
		}
	}

	return min
}

func getHighestScheduledBasalRate(basalRateSchedule *pump.BasalRateStartArray) *float64 {
	if basalRateSchedule == nil {
		return nil
	}
	var max *float64
	for _, b := range *basalRateSchedule {
		if b != nil && b.Rate != nil && (max == nil || *b.Rate > *max) {
			max = b.Rate
		}
	}
	return max
}
