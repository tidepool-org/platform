package guardrails

import (
	devices "github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
	"strconv"
)

func ValidateCarbohydrateRatioSchedule(carbohydrateRatioSchedule pump.CarbohydrateRatioStartArray, guardRail *devices.CarbohydrateRatioGuardRail, validator structure.Validator) {
	validValues := generateValidValuesFromAbsoluteBounds(guardRail.AbsoluteBounds)
	for i, carbRatio := range carbohydrateRatioSchedule {
		ValidateValueIfNotNil(carbRatio.Amount, validValues, validator.WithReference(strconv.Itoa(i)).WithReference("amount"))
	}
}
