package guardrails

import (
	devices "github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
	"strconv"
)

func ValidateInsulinSensitivitySchedule(insulinSensitivitySchedule pump.InsulinSensitivityStartArray, guardRail *devices.InsulinSensitivityGuardRail, validator structure.Validator) {
	validValues := generateValidValuesFromAbsoluteBounds(guardRail.AbsoluteBounds)
	for i, insulinSensitivity := range insulinSensitivitySchedule {
		ValidateValueIfNotNil(insulinSensitivity.Amount, validValues, validator.WithReference(strconv.Itoa(i)).WithReference("amount"))
	}
}
