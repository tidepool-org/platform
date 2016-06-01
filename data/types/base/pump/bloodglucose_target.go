package pump

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/bloodglucose"
)

// TODO: Consider moving this to common along with calculator/bloodglucose_target

type BloodGlucoseTarget struct {
	Low    *float64 `json:"low,omitempty" bson:"low,omitempty"`
	High   *float64 `json:"high,omitempty" bson:"high,omitempty"`
	Target *float64 `json:"target,omitempty" bson:"target,omitempty"`
	Range  *float64 `json:"range,omitempty" bson:"range,omitempty"`
	Start  *int     `json:"start,omitempty" bson:"start,omitempty"`
}

func NewBloodGlucoseTarget() *BloodGlucoseTarget {
	return &BloodGlucoseTarget{}
}

func (b *BloodGlucoseTarget) Parse(parser data.ObjectParser) {
	b.High = parser.ParseFloat("high")
	b.Low = parser.ParseFloat("low")
	b.Target = parser.ParseFloat("target")
	b.Range = parser.ParseFloat("range")
	b.Start = parser.ParseInteger("start")
}

func (b *BloodGlucoseTarget) Validate(validator data.Validator, units *string) {
	validator.ValidateFloatAsBloodGlucoseValue("low", b.Low).InRangeForUnits(units)
	validator.ValidateFloatAsBloodGlucoseValue("high", b.High).InRangeForUnits(units)
	validator.ValidateFloatAsBloodGlucoseValue("target", b.Target).InRangeForUnits(units)
	validator.ValidateFloatAsBloodGlucoseValue("range", b.Range).InRange(RangeForUnits(units))
	validator.ValidateInteger("start", b.Start).Exists().InRange(0, 86400000)

	if b.Low != nil {
		validator.ValidateFloat("high", b.High).GreaterThanOrEqualTo(*b.Low)
	} else if b.Target != nil {
		validator.ValidateFloat("high", b.High).GreaterThanOrEqualTo(*b.Target)
	}
}

func (b *BloodGlucoseTarget) Normalize(normalizer data.Normalizer, units *string) {
	bloodGlucoseNormalizer := normalizer.NormalizeBloodGlucose(units)
	b.Low = bloodGlucoseNormalizer.Value(b.Low)
	b.High = bloodGlucoseNormalizer.Value(b.High)
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
