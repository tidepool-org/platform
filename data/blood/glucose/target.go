package glucose

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

func (t *Target) Parse(parser data.ObjectParser) {
	t.Target = parser.ParseFloat("target")
	t.Range = parser.ParseFloat("range")
	t.Low = parser.ParseFloat("low")
	t.High = parser.ParseFloat("high")
}

func (t *Target) Validate(validator data.Validator, units *string) {
	if t.Target != nil && t.Range != nil {
		validator.ValidateFloat("target", t.Target).Exists().InRange(TargetRangeForUnits(units))
		validator.ValidateFloat("range", t.Range).Exists().InRange(RangeRangeForUnits(*t.Target, units))
		validator.ValidateFloat("low", t.Low).NotExists()
		validator.ValidateFloat("high", t.High).NotExists()
	} else if t.Target != nil && t.High != nil {
		validator.ValidateFloat("target", t.Target).Exists().InRange(TargetRangeForUnits(units))
		validator.ValidateFloat("range", t.Range).NotExists()
		validator.ValidateFloat("low", t.Low).NotExists()
		validator.ValidateFloat("high", t.High).Exists().InRange(HighRangeForUnits(*t.Target, units))
	} else if t.Target != nil {
		validator.ValidateFloat("target", t.Target).Exists().InRange(TargetRangeForUnits(units))
		validator.ValidateFloat("range", t.Range).NotExists()
		validator.ValidateFloat("low", t.Low).NotExists()
		validator.ValidateFloat("high", t.High).NotExists()
	} else if t.Low != nil && t.High != nil {
		validator.ValidateFloat("target", t.Target).NotExists()
		validator.ValidateFloat("range", t.Range).NotExists()
		validator.ValidateFloat("low", t.Low).Exists().InRange(LowRangeForUnits(units))
		validator.ValidateFloat("high", t.High).Exists().InRange(HighRangeForUnits(*t.Low, units))
	} else if t.Low != nil {
		validator.ValidateFloat("high", t.High).Exists()
	} else {
		validator.ValidateFloat("target", t.Target).Exists()
	}
}

func (t *Target) Normalize(normalizer data.Normalizer, units *string) {
	t.Target = NormalizeValueForUnits(t.Target, units)
	t.Range = NormalizeValueForUnits(t.Range, units)
	t.Low = NormalizeValueForUnits(t.Low, units)
	t.High = NormalizeValueForUnits(t.High, units)
}

func TargetRangeForUnits(units *string) (float64, float64) {
	return ValueRangeForUnits(units)
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

func LowRangeForUnits(units *string) (float64, float64) {
	return ValueRangeForUnits(units)
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
