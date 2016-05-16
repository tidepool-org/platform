package pump

import "github.com/tidepool-org/platform/pvn/data"

type BloodGlucoseTarget struct {
	Low    *float64 `json:"low" bson:"low"`
	High   *float64 `json:"high" bson:"high"`
	Target *float64 `json:"target" bson:"target"`
	Start  *int     `json:"start" bson:"start"`
	Range  *int     `json:"range" bson:"range"`
}

func NewBloodGlucoseTarget() *BloodGlucoseTarget {
	return &BloodGlucoseTarget{}
}

func (b *BloodGlucoseTarget) Parse(parser data.ObjectParser) {
	b.High = parser.ParseFloat("high")
	b.Low = parser.ParseFloat("low")
	b.Target = parser.ParseFloat("target")

	b.Start = parser.ParseInteger("start")
	b.Range = parser.ParseInteger("range")
}

func (b *BloodGlucoseTarget) Validate(validator data.Validator) {

	if b.High != nil {
		validator.ValidateFloat("low", b.Low).GreaterThanOrEqualTo(0.0).LessThanOrEqualTo(*b.High)
	} else {
		validator.ValidateFloat("low", b.Low).GreaterThanOrEqualTo(0.0).LessThanOrEqualTo(1000.0)
	}

	validator.ValidateFloat("target", b.Target).GreaterThanOrEqualTo(0.0).LessThanOrEqualTo(1000.0)

	if b.Low != nil {
		validator.ValidateFloat("high", b.High).GreaterThanOrEqualTo(*b.Low).LessThanOrEqualTo(1000.0)
	} else if b.Target != nil {
		validator.ValidateFloat("high", b.High).GreaterThanOrEqualTo(*b.Target).LessThanOrEqualTo(1000.0)
	}

	validator.ValidateInteger("range", b.Range).GreaterThanOrEqualTo(0).LessThanOrEqualTo(50)
	validator.ValidateInteger("start", b.Start).Exists().GreaterThanOrEqualTo(0).LessThanOrEqualTo(86400000)
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
