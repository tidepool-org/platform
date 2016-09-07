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
	"github.com/tidepool-org/platform/service"
)

type StandardStringAsBloodGlucoseUnits struct {
	context     data.Context
	reference   interface{}
	stringValue *string
}

func NewStandardStringAsBloodGlucoseUnits(context data.Context, reference interface{}, stringValue *string) *StandardStringAsBloodGlucoseUnits {
	if context == nil {
		return nil
	}

	standardStringAsBloodGlucoseUnits := &StandardStringAsBloodGlucoseUnits{
		context:     context,
		reference:   reference,
		stringValue: stringValue,
	}
	standardStringAsBloodGlucoseUnits.parse()
	return standardStringAsBloodGlucoseUnits
}

func (s *StandardStringAsBloodGlucoseUnits) Exists() data.BloodGlucoseUnits {
	if s.stringValue == nil {
		s.context.AppendError(s.reference, service.ErrorValueNotExists())
	}
	return s
}

func (s *StandardStringAsBloodGlucoseUnits) NotExists() data.BloodGlucoseUnits {
	if s.stringValue != nil {
		s.context.AppendError(s.reference, service.ErrorValueExists())
	}
	return s
}

func (s *StandardStringAsBloodGlucoseUnits) parse() {
	if s.stringValue != nil {
		switch *s.stringValue {
		case bloodglucose.Mmoll, bloodglucose.MmolL, bloodglucose.Mgdl, bloodglucose.MgdL:
		default:
			s.context.AppendError(s.reference, service.ErrorValueStringNotOneOf(*s.stringValue, []string{bloodglucose.Mmoll, bloodglucose.MmolL, bloodglucose.Mgdl, bloodglucose.MgdL}))
		}
	}
}
