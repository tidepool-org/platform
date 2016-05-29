package pump

import "github.com/tidepool-org/platform/data"

type BloodGlucoseTarget struct {
	Low    *float64 `json:"low,omitempty" bson:"low,omitempty"`
	High   *float64 `json:"high,omitempty" bson:"high,omitempty"`
	Target *float64 `json:"target,omitempty" bson:"target,omitempty"`
	Start  *int     `json:"start,omitempty" bson:"start,omitempty"`
	Range  *int     `json:"range,omitempty" bson:"range,omitempty"`
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

func (b *BloodGlucoseTarget) Validate(validator data.Validator, units *string) {
	validator.ValidateFloatAsBloodGlucoseValue("low", b.Low).InRangeForUnits(units)
	validator.ValidateFloatAsBloodGlucoseValue("high", b.High).InRangeForUnits(units)
	validator.ValidateFloatAsBloodGlucoseValue("target", b.Target).InRangeForUnits(units)

	if b.Low != nil {
		validator.ValidateFloat("high", b.High).GreaterThanOrEqualTo(*b.Low)
	} else if b.Target != nil {
		validator.ValidateFloat("high", b.High).GreaterThanOrEqualTo(*b.Target)
	}

	validator.ValidateInteger("range", b.Range).InRange(0, 50) // TODO: Isn't this units dependent?
	validator.ValidateInteger("start", b.Start).Exists().InRange(0, 86400000)
}

func (b *BloodGlucoseTarget) Normalize(normalizer data.Normalizer, units *string) {
	bloodGlucoseNormalizer := normalizer.NormalizeBloodGlucose(units)
	b.Low = bloodGlucoseNormalizer.Value(b.Low)
	b.High = bloodGlucoseNormalizer.Value(b.High)
	b.Target = bloodGlucoseNormalizer.Value(b.Target)
	// TODO: Don't we have to normalize range as well?
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
