package ketone

import (
	"math"

	"github.com/tidepool-org/platform/app"
)

const (
	MmolL string = "mmol/L"
	Mmoll string = "mmol/l"

	MmolLLowerLimit float64 = 0.0
	MmolLUpperLimit float64 = 10.0
)

func Units() []string {
	return []string{MmolL, Mmoll}
}

func ValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case MmolL, Mmoll:
			return MmolLLowerLimit, MmolLUpperLimit
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

func NormalizeUnits(units *string) *string {
	if units != nil {
		switch *units {
		case MmolL, Mmoll:
			return app.StringAsPointer(MmolL)
		}
	}
	return units
}

func NormalizeValueForUnits(value *float64, units *string) *float64 {
	return value
}
