package guardrails

import (
	devices "github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
	"strconv"
)

func ValidateBasalRateSchedule(basalRateSchedule pump.BasalRateStartArray, guardRail *devices.BasalRatesGuardRail, validator structure.Validator) {
	validValues := generateAllValidValues(guardRail.AbsoluteBounds)
	for i, basalRate := range basalRateSchedule {
		ValidateValueIfNotNil(basalRate.Rate, validValues, validator.WithReference(strconv.Itoa(i)).WithReference("rate"))
	}
}
