package pump

import (
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

func ValidateBloodGlucoseSuspendThreshold(suspendThreshold *float64, units *string, reference string, validator structure.Validator) {
	validator.Float64(reference, suspendThreshold).InRange(dataBloodGlucose.ValueRangeForUnits(units))
}
