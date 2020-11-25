package guardrails

import (
	devices "github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/structure"
)

func ValidateGlucoseSafetyLimit(glucoseSafetyLimit *float64, correctionRanges CorrectionRanges, guardRail *devices.GlucoseSafetyLimitGuardRail, validator structure.Validator) {
	validValues := generateGlucoseSafetyLimitValidValues(glucoseSafetyLimit, correctionRanges, guardRail)
	ValidateValueIfNotNil(glucoseSafetyLimit, validValues, validator)
}

func generateGlucoseSafetyLimitValidValues(glucoseSafetyLimit *float64, correctionRanges CorrectionRanges, guardRail *devices.GlucoseSafetyLimitGuardRail) []float64 {
	validValues := generateValidValuesFromAbsoluteBounds(guardRail.AbsoluteBounds)
	if bounds := correctionRanges.GetBounds(); bounds != nil {
		validValues = discardValuesLargerThan(validValues, bounds.Lower)
	}
	return validValues
}
