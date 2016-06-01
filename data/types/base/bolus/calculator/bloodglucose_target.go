package calculator

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/bloodglucose"
)

//NOTE: this is the matrix we are working to. Only animas at this stage
// animas: {`target`, `range`}
// insulet: {`target`, `high`}
// medtronic: {`low`, `high`}
// tandem: {`target`}

// TODO: Consider moving this to common along with pump/bloodglucose_target

type BloodGlucoseTarget struct {
	Target *float64 `json:"target,omitempty" bson:"target,omitempty"`
	Range  *float64 `json:"range,omitempty" bson:"range,omitempty"`
}

func NewBloodGlucoseTarget() *BloodGlucoseTarget {
	return &BloodGlucoseTarget{}
}

func (b *BloodGlucoseTarget) Parse(parser data.ObjectParser) {
	b.Target = parser.ParseFloat("target")
	b.Range = parser.ParseFloat("range")
}

func (b *BloodGlucoseTarget) Validate(validator data.Validator, units *string) {
	validator.ValidateFloatAsBloodGlucoseValue("target", b.Target).InRangeForUnits(units)
	validator.ValidateFloatAsBloodGlucoseValue("range", b.Range).InRange(RangeForUnits(units))
}

func (b *BloodGlucoseTarget) Normalize(normalizer data.Normalizer, units *string) {
	bloodGlucoseNormalizer := normalizer.NormalizeBloodGlucose(units)
	b.Target = bloodGlucoseNormalizer.Value(b.Target)
	b.Range = bloodGlucoseNormalizer.Value(b.Range)
}

func ParseBloodGlucoseTarget(parser data.ObjectParser) *BloodGlucoseTarget {
	var bloodGlucoseTarget *BloodGlucoseTarget
	if parser.Object() != nil {
		bloodGlucoseTarget = NewBloodGlucoseTarget()
		bloodGlucoseTarget.Parse(parser)
	}
	return bloodGlucoseTarget
}

func RangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case bloodglucose.MmolL, bloodglucose.Mmoll:
			return 0.0, 3.0
		case bloodglucose.MgdL, bloodglucose.Mgdl:
			return 0.0, 50.0
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
