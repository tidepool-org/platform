package ketone

import (
	"math"

	"github.com/tidepool-org/platform/pointer"
)

const (
	MmolL = "mmol/L"
	Mmoll = "mmol/l"

	MmolLMinimum = 0.0
	MmolLMaximum = 10.0
)

func Units() []string {
	return []string{MmolL, Mmoll}
}

func ValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case MmolL, Mmoll:
			return MmolLMinimum, MmolLMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

func NormalizeUnits(units *string) *string {
	if units != nil {
		switch *units {
		case MmolL, Mmoll:
			return pointer.FromString(MmolL)
		}
	}
	return units
}

func NormalizeValueForUnits(value *float64, units *string) *float64 {
	return value
}
