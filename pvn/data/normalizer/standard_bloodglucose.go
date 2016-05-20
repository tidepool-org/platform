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
	return &common.MmolL
}

func (s *StandardBloodGlucose) NormalizeValue(value *float64) *float64 {

	if value == nil {
		s.context.AppendError(s.reference, types.ErrorValueMissing())
		return nil
	}

	switch s.units {
	case &common.Mmoll, &common.MmolL:
		return value
	default:
		converted := *value / common.MgdlToMmolConversion
		return &converted
	}
}

func (s *StandardBloodGlucose) NormalizeUnitsAndValue(value *float64) (*string, *float64) {
	return s.NormalizeUnits(), s.NormalizeValue(value)
}
