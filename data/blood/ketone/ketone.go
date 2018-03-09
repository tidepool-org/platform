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

	MmolLPrecisionFactor float64 = 100000.0
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
			return pointer.String(MmolL)
		}
	}
	return units
}

func NormalizeValueForUnits(value *float64, units *string) *float64 {
	if value != nil && units != nil {
		switch *units {
		case MmolL, Mmoll:
			// TODO: Normalize mmol/L values to standard precision
			// return NormalizePrecisionForUnits(value, units)
		}
	}
	return value
}

func NormalizePrecisionForUnits(value *float64, units *string) *float64 {
	if value != nil && units != nil {
		switch *units {
		case MmolL, Mmoll:
			return pointer.Float64(float64(int(*value*MmolLPrecisionFactor+0.5)) / MmolLPrecisionFactor)
		}
	}
	return value
}
