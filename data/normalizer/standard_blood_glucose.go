package normalizer

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import "github.com/tidepool-org/platform/data/bloodglucose"

type StandardBloodGlucose struct {
	units *string
}

func NewStandardBloodGlucose(units *string) *StandardBloodGlucose {
	return &StandardBloodGlucose{
		units: units,
	}
}

func (s *StandardBloodGlucose) Units() *string {
	if s.units != nil {
		switch *s.units {
		case bloodglucose.Mmoll, bloodglucose.MmolL, bloodglucose.Mgdl, bloodglucose.MgdL:
			units := bloodglucose.MmolL
			return &units
		}
	}
	return nil
}

func (s *StandardBloodGlucose) Value(value *float64) *float64 {
	if s.units != nil && value != nil {
		converted := bloodglucose.ConvertValue(*value, *s.units, bloodglucose.MmolL)
		value = &converted
	}
	return value
}

func (s *StandardBloodGlucose) UnitsAndValue(value *float64) (*string, *float64) {
	return s.Units(), s.Value(value)
}
