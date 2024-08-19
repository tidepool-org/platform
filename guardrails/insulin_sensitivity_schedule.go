package guardrails

import (
	devices "github.com/tidepool-org/devices/api"

	structureValidator "github.com/tidepool-org/platform/structure/validator"

	"strconv"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
)

func ValidateInsulinSensitivitySchedule(insulinSensitivitySchedule pump.InsulinSensitivityStartArray, guardRail *devices.InsulinSensitivityGuardRail, validator structure.Validator) {
	if guardRail.MaxSegments != nil && len(insulinSensitivitySchedule) > int(*guardRail.MaxSegments) {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(len(insulinSensitivitySchedule), int(*guardRail.MaxSegments)))
	}
	validValues := generateValidValuesFromAbsoluteBounds(guardRail.AbsoluteBounds)
	for i, insulinSensitivity := range insulinSensitivitySchedule {
		ValidateValueIfNotNil(insulinSensitivity.Amount, validValues, validator.WithReference(strconv.Itoa(i)).WithReference("amount"))
	}
}
