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
	Target      *float64 `json:"target" bson:"target"`
	Range       *int     `json:"range" bson:"range"`
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
	case &common.Mmoll, &common.MmolL:
		validator.ValidateFloat("target", b.Target).InRange(common.MmolLFromValue, common.MmolLToValue)
	default:
		validator.ValidateFloat("target", b.Target).InRange(common.MgdLFromValue, common.MgdLToValue)
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
