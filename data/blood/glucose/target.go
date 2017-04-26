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

func (b *Target) Parse(parser data.ObjectParser) {
	b.Target = parser.ParseFloat("target")
	b.Range = parser.ParseFloat("range")
	b.Low = parser.ParseFloat("low")
	b.High = parser.ParseFloat("high")
}

func (b *Target) Validate(validator data.Validator, units *string) {
	if b.Target != nil && b.Range != nil {
		validator.ValidateFloat("target", b.Target).Exists().InRange(TargetRangeForUnits(units))
		validator.ValidateFloat("range", b.Range).Exists().InRange(RangeRangeForUnits(*b.Target, units))
		validator.ValidateFloat("low", b.Low).NotExists()
		validator.ValidateFloat("high", b.High).NotExists()
	} else if b.Target != nil && b.High != nil {
		validator.ValidateFloat("target", b.Target).Exists().InRange(TargetRangeForUnits(units))
		validator.ValidateFloat("range", b.Range).NotExists()
		validator.ValidateFloat("low", b.Low).NotExists()
		validator.ValidateFloat("high", b.High).Exists().InRange(HighRangeForUnits(*b.Target, units))
	} else if b.Target != nil {
		validator.ValidateFloat("target", b.Target).Exists().InRange(TargetRangeForUnits(units))
		validator.ValidateFloat("range", b.Range).NotExists()
		validator.ValidateFloat("low", b.Low).NotExists()
		validator.ValidateFloat("high", b.High).NotExists()
	} else if b.Low != nil && b.High != nil {
		validator.ValidateFloat("target", b.Target).NotExists()
		validator.ValidateFloat("range", b.Range).NotExists()
		validator.ValidateFloat("low", b.Low).Exists().InRange(LowRangeForUnits(units))
		validator.ValidateFloat("high", b.High).Exists().InRange(HighRangeForUnits(*b.Low, units))
	} else if b.Low != nil {
		validator.ValidateFloat("high", b.High).Exists()
	} else {
		validator.ValidateFloat("target", b.Target).Exists()
	}
}

func (b *Target) Normalize(normalizer data.Normalizer, units *string) {
	b.Target = NormalizeValueForUnits(b.Target, units)
	b.Range = NormalizeValueForUnits(b.Range, units)
	b.Low = NormalizeValueForUnits(b.Low, units)
	b.High = NormalizeValueForUnits(b.High, units)
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
