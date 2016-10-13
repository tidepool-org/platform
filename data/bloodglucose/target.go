package bloodglucose

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"math"

	"github.com/tidepool-org/platform/data"
)

type Target struct {
	Target *float64 `json:"target,omitempty" bson:"target,omitempty"`
	Range  *float64 `json:"range,omitempty" bson:"range,omitempty"`
	Low    *float64 `json:"low,omitempty" bson:"low,omitempty"`
	High   *float64 `json:"high,omitempty" bson:"high,omitempty"`
}

func ParseTarget(parser data.ObjectParser) *Target {
	var target *Target
	if parser.Object() != nil {
		target = NewTarget()
		target.Parse(parser)
		parser.ProcessNotParsed()
	}
	return target
}

func NewTarget() *Target {
	return &Target{}
}

func (b *Target) Parse(parser data.ObjectParser) {
	b.Target = parser.ParseFloat("target")
	b.Range = parser.ParseFloat("range")
	b.Low = parser.ParseFloat("low")
	b.High = parser.ParseFloat("high")
}

func (b *Target) Validate(validator data.Validator, units *string) {
	if b.Target != nil && b.Range != nil {
		validator.ValidateFloatAsBloodGlucoseValue("target", b.Target).Exists().InRangeForUnits(units)
		validator.ValidateFloatAsBloodGlucoseValue("range", b.Range).Exists().InRange(RangeRangeForUnits(*b.Target, units))
		validator.ValidateFloatAsBloodGlucoseValue("low", b.Low).NotExists()
		validator.ValidateFloatAsBloodGlucoseValue("high", b.High).NotExists()
	} else if b.Target != nil && b.High != nil {
		validator.ValidateFloatAsBloodGlucoseValue("target", b.Target).Exists().InRangeForUnits(units)
		validator.ValidateFloatAsBloodGlucoseValue("range", b.Range).NotExists()
		validator.ValidateFloatAsBloodGlucoseValue("low", b.Low).NotExists()
		validator.ValidateFloatAsBloodGlucoseValue("high", b.High).Exists().InRange(HighRangeForUnits(*b.Target, units))
	} else if b.Target != nil {
		validator.ValidateFloatAsBloodGlucoseValue("target", b.Target).Exists().InRangeForUnits(units)
		validator.ValidateFloatAsBloodGlucoseValue("range", b.Range).NotExists()
		validator.ValidateFloatAsBloodGlucoseValue("low", b.Low).NotExists()
		validator.ValidateFloatAsBloodGlucoseValue("high", b.High).NotExists()
	} else if b.Low != nil && b.High != nil {
		validator.ValidateFloatAsBloodGlucoseValue("target", b.Target).NotExists()
		validator.ValidateFloatAsBloodGlucoseValue("range", b.Range).NotExists()
		validator.ValidateFloatAsBloodGlucoseValue("low", b.Low).Exists().InRangeForUnits(units)
		validator.ValidateFloatAsBloodGlucoseValue("high", b.High).Exists().InRange(HighRangeForUnits(*b.Low, units))
	} else if b.Low != nil {
		validator.ValidateFloatAsBloodGlucoseValue("high", b.High).Exists()
	} else {
		validator.ValidateFloatAsBloodGlucoseValue("target", b.Target).Exists()
	}
}

func (b *Target) Normalize(normalizer data.Normalizer, units *string) {
	bloodGlucoseNormalizer := normalizer.NormalizeBloodGlucose(units)
	b.Target = bloodGlucoseNormalizer.Value(b.Target)
	b.Range = bloodGlucoseNormalizer.Value(b.Range)
	b.Low = bloodGlucoseNormalizer.Value(b.Low)
	b.High = bloodGlucoseNormalizer.Value(b.High)
}

func RangeRangeForUnits(target float64, units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case MmolL, Mmoll:
			if target >= MmolLLowerLimit && target <= MmolLUpperLimit {
				return 0.0, math.Min(target-MmolLLowerLimit, MmolLUpperLimit-target)
			}
		case MgdL, Mgdl:
			if target >= MgdLLowerLimit && target <= MgdLUpperLimit {
				return 0.0, math.Min(target-MgdLLowerLimit, MgdLUpperLimit-target)
			}
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

func HighRangeForUnits(low float64, units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case MmolL, Mmoll:
			if low >= MmolLLowerLimit && low <= MmolLUpperLimit {
				return low, MmolLUpperLimit
			}
		case MgdL, Mgdl:
			if low >= MgdLLowerLimit && low <= MgdLUpperLimit {
				return low, MgdLUpperLimit
			}
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
