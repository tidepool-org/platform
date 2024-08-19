package guardrails

import (
	devices "github.com/tidepool-org/devices/api"

	structureValidator "github.com/tidepool-org/platform/structure/validator"

	"strconv"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
)

func ValidateCarbohydrateRatioSchedule(carbohydrateRatioSchedule pump.CarbohydrateRatioStartArray, guardRail *devices.CarbohydrateRatioGuardRail, validator structure.Validator) {
	if guardRail.MaxSegments != nil && len(carbohydrateRatioSchedule) > int(*guardRail.MaxSegments) {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(len(carbohydrateRatioSchedule), int(*guardRail.MaxSegments)))
	}
	validValues := generateValidValuesFromAbsoluteBounds(guardRail.AbsoluteBounds)
	for i, carbRatio := range carbohydrateRatioSchedule {
		ValidateValueIfNotNil(carbRatio.Amount, validValues, validator.WithReference(strconv.Itoa(i)).WithReference("amount"))
	}
}
