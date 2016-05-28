package validator

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/bloodglucose"
)

type StandardFloatAsBloodGlucoseValue struct {
	context    data.Context
	reference  interface{}
	floatValue *float64
}

func NewStandardFloatAsBloodGlucoseValue(context data.Context, reference interface{}, floatValue *float64) *StandardFloatAsBloodGlucoseValue {
	if context == nil {
		return nil
	}

	standardFloatAsBloodGlucoseValue := &StandardFloatAsBloodGlucoseValue{
		context:    context,
		reference:  reference,
		floatValue: floatValue,
	}
	return standardFloatAsBloodGlucoseValue
}

func (s *StandardFloatAsBloodGlucoseValue) Exists() data.BloodGlucoseValue {
	if s.floatValue == nil {
		s.context.AppendError(s.reference, ErrorValueDoesNotExist())
	}
	return s
}

func (s *StandardFloatAsBloodGlucoseValue) InRange(lowerLimit float64, upperLimit float64) data.BloodGlucoseValue {
	if s.floatValue != nil {
		if *s.floatValue < lowerLimit || *s.floatValue > upperLimit {
			s.context.AppendError(s.reference, ErrorFloatNotInRange(*s.floatValue, lowerLimit, upperLimit))
		}
	}
	return s
}

func (s *StandardFloatAsBloodGlucoseValue) InRangeForUnits(units *string) data.BloodGlucoseValue {
	if units != nil {
		switch *units {
		case bloodglucose.MmolL, bloodglucose.Mmoll:
			return s.InRange(bloodglucose.MmolLLowerLimit, bloodglucose.MmolLUpperLimit)
		case bloodglucose.MgdL, bloodglucose.Mgdl:
			return s.InRange(bloodglucose.MgdLLowerLimit, bloodglucose.MgdLUpperLimit)
		}
	}
	return s
}
