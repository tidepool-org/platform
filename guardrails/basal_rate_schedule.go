package guardrails

import (
	devices "github.com/tidepool-org/devices/api"

	structureValidator "github.com/tidepool-org/platform/structure/validator"

	"strconv"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
)

func ValidateBasalRateSchedule(basalRateSchedule pump.BasalRateStartArray, guardRail *devices.BasalRatesGuardRail, validator structure.Validator) {
	if guardRail.MaxSegments != nil && len(basalRateSchedule) > int(*guardRail.MaxSegments) {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(len(basalRateSchedule), int(*guardRail.MaxSegments)))
	}
	validValues := generateAllValidValues(guardRail.AbsoluteBounds)
	for i, basalRate := range basalRateSchedule {
		ValidateValueIfNotNil(basalRate.Rate, validValues, validator.WithReference(strconv.Itoa(i)).WithReference("rate"))
	}
}
