package calculator

import "github.com/tidepool-org/platform/pvn/data"

// animas: {`target`, `range`}
// insulet: {`target`, `high`}
// medtronic: {`low`, `high`}
// tandem: {`target`}

type BloodGlucoseTarget struct {
	Target *float64 `json:"target" bson:"target"`
	Range  *int     `json:"range" bson:"range"`
}

func NewBloodGlucoseTarget() *BloodGlucoseTarget {
	return &BloodGlucoseTarget{}
}

func (b *BloodGlucoseTarget) Parse(parser data.ObjectParser) {
	b.Target = parser.ParseFloat("target")
	b.Range = parser.ParseInteger("range")
}

func (b *BloodGlucoseTarget) Validate(validator data.Validator) {
	validator.ValidateFloat("target", b.Target).GreaterThanOrEqualTo(0.0).LessThanOrEqualTo(1000.0)
	validator.ValidateInteger("range", b.Range).GreaterThanOrEqualTo(0).LessThanOrEqualTo(50)
}

func (b *BloodGlucoseTarget) Normalize(normalizer data.Normalizer) {
}

func ParseBloodGlucoseTarget(parser data.ObjectParser) *BloodGlucoseTarget {
	var bloodGlucoseTarget *BloodGlucoseTarget
	if parser.Object() != nil {
		bloodGlucoseTarget = NewBloodGlucoseTarget()
		bloodGlucoseTarget.Parse(parser)
	}
	return bloodGlucoseTarget
}
