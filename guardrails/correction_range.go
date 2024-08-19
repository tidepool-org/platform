package guardrails

import (
	"math"
	"strconv"

	structureValidator "github.com/tidepool-org/platform/structure/validator"

	devices "github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
)

func ValidateBloodGlucoseTargetSchedule(bloodGlucoseTargetSchedule pump.BloodGlucoseTargetStartArray, glucoseSafetyLimit *float64, guardRail *devices.CorrectionRangeGuardRail, validator structure.Validator) {
	if guardRail.MaxSegments != nil && len(bloodGlucoseTargetSchedule) > int(*guardRail.MaxSegments) {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(len(bloodGlucoseTargetSchedule), int(*guardRail.MaxSegments)))
	}
	for i, bloodGlucoseTargetStart := range bloodGlucoseTargetSchedule {
		ValidateBloodGlucoseTarget(bloodGlucoseTargetStart.Target, glucoseSafetyLimit, guardRail, validator.WithReference(strconv.Itoa(i)))
	}
}

func ValidateBloodGlucoseTarget(bloodGlucoseTarget glucose.Target, glucoseSafetyLimit *float64, guardRail *devices.CorrectionRangeGuardRail, validator structure.Validator) {
	validValues := generateValidValuesFromAbsoluteBounds(guardRail.AbsoluteBounds)
	if glucoseSafetyLimit != nil {
		validValues = discardValuesSmallerThan(validValues, math.Max(validValues[0], *glucoseSafetyLimit))
	}
	if bounds := bloodGlucoseTarget.GetBounds(); bounds != nil {
		ValidateValueIfNotNil(&bounds.Lower, validValues, validator.WithReference("low"))
		ValidateValueIfNotNil(&bounds.Upper, validValues, validator.WithReference("high"))
	}
}

type CorrectionRanges struct {
	Schedule         *pump.BloodGlucoseTargetStartArray
	Preprandial      *glucose.Target
	PhysicalActivity *glucose.Target
}

func (c CorrectionRanges) GetBounds() *glucose.Bounds {
	targets := make(pump.BloodGlucoseTargetStartArray, 0)
	if c.Schedule != nil {
		for _, target := range *c.Schedule {
			targets = append(targets, target)
		}
	}
	if c.Preprandial != nil {
		targets = append(targets, &pump.BloodGlucoseTargetStart{
			Target: *c.Preprandial,
		})
	}
	if c.PhysicalActivity != nil {
		targets = append(targets, &pump.BloodGlucoseTargetStart{
			Target: *c.PhysicalActivity,
		})
	}

	return targets.GetBounds()
}
