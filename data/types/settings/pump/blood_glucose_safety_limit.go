package pump

import (
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

func ValidateBloodGlucoseSafetyLimit(safetyLimit *float64, units *string, reference string, validator structure.Validator) {
	validator.Float64(reference, safetyLimit).InRange(dataBloodGlucose.ValueRangeForUnits(units))
}
