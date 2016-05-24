package calculator

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/common/bloodglucose"
)

//NOTE: this is the matrix we are working to. Only animas at this stage
// animas: {`target`, `range`}
// insulet: {`target`, `high`}
// medtronic: {`low`, `high`}
// tandem: {`target`}

type BloodGlucoseTarget struct {
	Target      *float64 `json:"target,omitempty" bson:"target,omitempty"`
	Range       *int     `json:"range,omitempty" bson:"range,omitempty"`
	targetUnits *string
}

func NewBloodGlucoseTarget() *BloodGlucoseTarget {
	return &BloodGlucoseTarget{}
}

func (b *BloodGlucoseTarget) Parse(parser data.ObjectParser) {
	b.Target = parser.ParseFloat("target")
	b.Range = parser.ParseInteger("range")
}

func (b *BloodGlucoseTarget) Validate(validator data.Validator) {
	switch b.targetUnits {
	case &bloodglucose.Mmoll, &bloodglucose.MmolL:
		validator.ValidateFloat("target", b.Target).InRange(bloodglucose.MmolLFromValue, bloodglucose.MmolLToValue)
	default:
		validator.ValidateFloat("target", b.Target).InRange(bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue)
	}

	validator.ValidateInteger("range", b.Range).InRange(0, 50)
}

func (b *BloodGlucoseTarget) Normalize(normalizer data.Normalizer) {
	b.Target = normalizer.NormalizeBloodGlucose("target", b.targetUnits).NormalizeValue(b.Target)
}

func ParseBloodGlucoseTarget(parser data.ObjectParser) *BloodGlucoseTarget {
	var bloodGlucoseTarget *BloodGlucoseTarget
	if parser.Object() != nil {
		bloodGlucoseTarget = NewBloodGlucoseTarget()
		bloodGlucoseTarget.Parse(parser)
	}
	return bloodGlucoseTarget
}
