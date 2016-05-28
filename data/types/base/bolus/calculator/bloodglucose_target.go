package calculator

import "github.com/tidepool-org/platform/data"

//NOTE: this is the matrix we are working to. Only animas at this stage
// animas: {`target`, `range`}
// insulet: {`target`, `high`}
// medtronic: {`low`, `high`}
// tandem: {`target`}

type BloodGlucoseTarget struct {
	Target *float64 `json:"target,omitempty" bson:"target,omitempty"`
	Range  *int     `json:"range,omitempty" bson:"range,omitempty"`
}

func NewBloodGlucoseTarget() *BloodGlucoseTarget {
	return &BloodGlucoseTarget{}
}

func (b *BloodGlucoseTarget) Parse(parser data.ObjectParser) {
	b.Target = parser.ParseFloat("target")
	b.Range = parser.ParseInteger("range")
}

func (b *BloodGlucoseTarget) Validate(validator data.Validator, units *string) {
	validator.ValidateFloatAsBloodGlucoseValue("target", b.Target).InRangeForUnits(units)
	validator.ValidateInteger("range", b.Range).InRange(0, 50) // TODO: Isn't range relative to units? 0-50 doesn't make sense for mmoll
}

func (b *BloodGlucoseTarget) Normalize(normalizer data.Normalizer, units *string) {
	b.Target = normalizer.NormalizeBloodGlucose(units).Value(b.Target)
	// TODO: Don't we have to normalize the range as it should be relative to the units?
}

func ParseBloodGlucoseTarget(parser data.ObjectParser) *BloodGlucoseTarget {
	var bloodGlucoseTarget *BloodGlucoseTarget
	if parser.Object() != nil {
		bloodGlucoseTarget = NewBloodGlucoseTarget()
		bloodGlucoseTarget.Parse(parser)
	}
	return bloodGlucoseTarget
}
