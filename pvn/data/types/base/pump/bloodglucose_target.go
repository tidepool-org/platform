package pump

import "github.com/tidepool-org/platform/pvn/data"

type BloodGlucoseTarget struct {
	Low   *float64 `json:"low" bson:"low"`
	High  *float64 `json:"high" bson:"high"`
	Start *int     `json:"start" bson:"start"`
}

func NewBloodGlucoseTarget() *BloodGlucoseTarget {
	return &BloodGlucoseTarget{}
}

func (b *BloodGlucoseTarget) Parse(parser data.ObjectParser) {
	b.High = parser.ParseFloat("high")
	b.Low = parser.ParseFloat("low")
	b.Start = parser.ParseInteger("start")
}

func (b *BloodGlucoseTarget) Validate(validator data.Validator) {
	validator.ValidateFloat("low", b.Low).Exists().GreaterThanOrEqualTo(0.0).LessThanOrEqualTo(*b.High)
	validator.ValidateFloat("high", b.High).Exists().GreaterThanOrEqualTo(*b.Low).LessThanOrEqualTo(1000.0)
	validator.ValidateInteger("start", b.Start).Exists().GreaterThanOrEqualTo(0)
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

func ParseBloodGlucoseTargetArray(parser data.ArrayParser) *[]*BloodGlucoseTarget {
	var bloodGlucoseTargetArray *[]*BloodGlucoseTarget
	if parser.Array() != nil {
		bloodGlucoseTargetArray = &[]*BloodGlucoseTarget{}
		for index := range *parser.Array() {
			*bloodGlucoseTargetArray = append(*bloodGlucoseTargetArray, ParseBloodGlucoseTarget(parser.NewChildObjectParser(index)))
		}
	}
	return bloodGlucoseTargetArray
}
