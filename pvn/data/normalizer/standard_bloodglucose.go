package normalizer

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types"
	"github.com/tidepool-org/platform/pvn/data/types/common/bloodglucose"
)

type StandardBloodGlucose struct {
	context   data.Context
	reference interface{}
	units     *string
}

func NewStandardBloodGlucose(context data.Context, reference interface{}, units *string) *StandardBloodGlucose {
	if context == nil {
		return nil
	}

	return &StandardBloodGlucose{
		context:   context,
		reference: reference,
		units:     units,
	}
}

func (s *StandardBloodGlucose) NormalizeUnits() *string {
	return &bloodglucose.MmolL
}

func (s *StandardBloodGlucose) NormalizeValue(value *float64) *float64 {

	// TODO: This should be a system error (not a parsing error)
	if value == nil {
		s.context.AppendError(s.reference, types.ErrorValueMissing())
		return nil
	}

	switch s.units {
	case &bloodglucose.Mmoll, &bloodglucose.MmolL:
		return value
	default:
		converted := *value / bloodglucose.MgdlToMmolConversion
		return &converted
	}
}

func (s *StandardBloodGlucose) NormalizeUnitsAndValue(value *float64) (*string, *float64) {
	// TODO: This could yield strange results if the value is null, (units will be normalized, but not value)
	return s.NormalizeUnits(), s.NormalizeValue(value)
}
