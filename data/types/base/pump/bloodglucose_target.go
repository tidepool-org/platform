package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/common/bloodglucose"
)

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

	if units == nil {
		return
	}

	lowBgUpperLimit := b.High
	highBgLowerLimit := b.Low

	switch units {
	case &bloodglucose.Mmoll, &bloodglucose.MmolL:

		if lowBgUpperLimit == nil {
			lowBgUpperLimit = &bloodglucose.MmolLToValue
		}
		if highBgLowerLimit == nil {
			if b.Target != nil {
				highBgLowerLimit = b.Target
			} else {
				highBgLowerLimit = &bloodglucose.MmolLFromValue
			}
		}

		validator.ValidateFloat("low", b.Low).InRange(bloodglucose.MmolLFromValue, *lowBgUpperLimit)
		validator.ValidateFloat("target", b.Target).InRange(bloodglucose.MmolLFromValue, bloodglucose.MmolLToValue)
		validator.ValidateFloat("high", b.High).GreaterThan(*highBgLowerLimit).LessThanOrEqualTo(bloodglucose.MmolLToValue)

	default:

		if lowBgUpperLimit == nil {
			lowBgUpperLimit = &bloodglucose.MgdLToValue
		}
		if highBgLowerLimit == nil {
			if b.Target != nil {
				highBgLowerLimit = b.Target
			} else {
				highBgLowerLimit = &bloodglucose.MgdLFromValue
			}
		}
		validator.ValidateFloat("low", b.Low).InRange(bloodglucose.MgdLFromValue, *lowBgUpperLimit)
		validator.ValidateFloat("target", b.Target).InRange(bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue)
		validator.ValidateFloat("high", b.High).GreaterThan(*highBgLowerLimit).LessThanOrEqualTo(bloodglucose.MgdLToValue)
	}

	validator.ValidateInteger("range", b.Range).InRange(0, 50)
	validator.ValidateInteger("start", b.Start).Exists().InRange(0, 86400000)
}

func (b *BloodGlucoseTarget) Normalize(normalizer data.Normalizer, units *string) {
	if units == nil {
		return
	}
	if b.Low != nil {
		b.Low = normalizer.NormalizeBloodGlucose("low", units).NormalizeValue(b.Low)
	}
	if b.High != nil {
		b.High = normalizer.NormalizeBloodGlucose("high", units).NormalizeValue(b.High)
	}
	if b.Target != nil {
		b.Target = normalizer.NormalizeBloodGlucose("target", units).NormalizeValue(b.Target)
	}
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
