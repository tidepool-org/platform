package guardrails

import (
	devices "github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
)

func ValidateBolusAmountMaximum(bolusAmountMaximum pump.BolusAmountMaximum, guardRail *devices.BolusAmountMaximumGuardRail, validator structure.Validator) {
	validValues := generateValidValuesFromAbsoluteBounds(guardRail.AbsoluteBounds)
	ValidateValueIfNotNil(bolusAmountMaximum.Value, validValues, validator.WithReference("value"))
}
