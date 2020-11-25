package guardrails

import (
	devices "github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
	"math"
	"strconv"
)

func ValidateBloodGlucoseTargetSchedule(bloodGlucoseTargetSchedule pump.BloodGlucoseTargetStartArray, glucoseSafetyLimit *float64, guardRail *devices.CorrectionRangeGuardRail, validator structure.Validator) {
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
	schedule *pump.BloodGlucoseTargetStartArray
	premeal  *glucose.Target
	workout  *glucose.Target
}

func (c *CorrectionRanges) GetBounds() *glucose.Bounds {
	allBounds := c.getNonEmptyBounds()
	if len(allBounds) == 0 {
		return nil
	}

	bounds := &allBounds[0]
	for _, b := range allBounds {
		if b.Lower < bounds.Lower {
			bounds.Lower = b.Lower
		}
		if b.Upper > bounds.Upper {
			bounds.Upper = b.Upper
		}
	}
	return bounds
}

func (c *CorrectionRanges) getNonEmptyBounds() []glucose.Bounds {
	bounds := make([]glucose.Bounds, 0)
	if c.schedule != nil {
		if b := c.schedule.GetBounds(); b != nil {
			bounds = append(bounds, *b)
		}
	}
	if c.premeal != nil && c.premeal.GetBounds() != nil{
		bounds = append(bounds, *c.premeal.GetBounds())
	}
	if c.premeal != nil && c.premeal.GetBounds() != nil{
		bounds = append(bounds, *c.premeal.GetBounds())
	}
	return bounds
}
