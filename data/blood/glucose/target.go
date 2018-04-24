package glucose

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type Target struct {
	High   *float64 `json:"high,omitempty" bson:"high,omitempty"`
	Low    *float64 `json:"low,omitempty" bson:"low,omitempty"`
	Range  *float64 `json:"range,omitempty" bson:"range,omitempty"`
	Target *float64 `json:"target,omitempty" bson:"target,omitempty"`
}

func ParseTarget(parser data.ObjectParser) *Target {
	if parser.Object() == nil {
		return nil
	}
	target := NewTarget()
	target.Parse(parser)
	parser.ProcessNotParsed()
	return target
}

func NewTarget() *Target {
	return &Target{}
}

func (t *Target) Parse(parser data.ObjectParser) {
	t.High = parser.ParseFloat("high")
	t.Low = parser.ParseFloat("low")
	t.Range = parser.ParseFloat("range")
	t.Target = parser.ParseFloat("target")
}

func (t *Target) Validate(validator structure.Validator, units *string) {
	if t.Target != nil && t.Range != nil {
		validator.Float64("high", t.High).NotExists()
		validator.Float64("low", t.Low).NotExists()
		validator.Float64("range", t.Range).Exists().InRange(RangeRangeForUnits(*t.Target, units))
		validator.Float64("target", t.Target).Exists().InRange(TargetRangeForUnits(units))
	} else if t.Target != nil && t.High != nil {
		validator.Float64("high", t.High).Exists().InRange(HighRangeForUnits(*t.Target, units))
		validator.Float64("low", t.Low).NotExists()
		validator.Float64("range", t.Range).NotExists()
		validator.Float64("target", t.Target).Exists().InRange(TargetRangeForUnits(units))
	} else if t.Target != nil {
		validator.Float64("high", t.High).NotExists()
		validator.Float64("low", t.Low).NotExists()
		validator.Float64("range", t.Range).NotExists()
		validator.Float64("target", t.Target).Exists().InRange(TargetRangeForUnits(units))
	} else if t.High != nil && t.Low != nil {
		validator.Float64("high", t.High).Exists().InRange(HighRangeForUnits(*t.Low, units))
		validator.Float64("low", t.Low).Exists().InRange(LowRangeForUnits(units))
		validator.Float64("range", t.Range).NotExists()
		validator.Float64("target", t.Target).NotExists()
	} else if t.Low != nil {
		validator.Float64("high", t.High).Exists()
	} else {
		validator.Float64("target", t.Target).Exists()
	}
}

func (t *Target) Normalize(normalizer data.Normalizer, units *string) {
	if normalizer.Origin() == structure.OriginExternal {
		t.High = NormalizeValueForUnits(t.High, units)
		t.Low = NormalizeValueForUnits(t.Low, units)
		t.Range = NormalizeValueForUnits(t.Range, units)
		t.Target = NormalizeValueForUnits(t.Target, units)
	}
}

func HighRangeForUnits(low float64, units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case MmolL, Mmoll:
			if low >= MmolLMinimum && low <= MmolLMaximum {
				return low, MmolLMaximum
			}
		case MgdL, Mgdl:
			if low >= MgdLMinimum && low <= MgdLMaximum {
				return low, MgdLMaximum
			}
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

func LowRangeForUnits(units *string) (float64, float64) {
	return ValueRangeForUnits(units)
}

func RangeRangeForUnits(target float64, units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case MmolL, Mmoll:
			if target >= MmolLMinimum && target <= MmolLMaximum {
				return 0.0, math.Min(target-MmolLMinimum, MmolLMaximum-target)
			}
		case MgdL, Mgdl:
			if target >= MgdLMinimum && target <= MgdLMaximum {
				return 0.0, math.Min(target-MgdLMinimum, MgdLMaximum-target)
			}
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

func TargetRangeForUnits(units *string) (float64, float64) {
	return ValueRangeForUnits(units)
}
