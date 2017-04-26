package glucose

import (
	"math"

	"github.com/tidepool-org/platform/app"
)

const (
	MmolL = "mmol/L"
	Mmoll = "mmol/l"

	MgdL = "mg/dL"
	Mgdl = "mg/dl"

	MmolLLowerLimit float64 = 0.0
	MmolLUpperLimit float64 = 55.0

	MgdLLowerLimit float64 = 0.0
	MgdLUpperLimit float64 = 1000.0

	MmolLToMgdLConversionFactor float64 = 18.01559
	MmolLToMgdLPrecisionFactor  float64 = 100000.0
)

func Units() []string {
	return []string{MmolL, Mmoll, MgdL, Mgdl}
}

func ValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case MmolL, Mmoll:
			return MmolLLowerLimit, MmolLUpperLimit
		case MgdL, Mgdl:
			return MgdLLowerLimit, MgdLUpperLimit
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

func NormalizeUnits(units *string) *string {
	if units != nil {
		switch *units {
		case MmolL, Mmoll, MgdL, Mgdl:
			return app.StringAsPointer(MmolL)
		}
	}
	return units
}

func NormalizeValueForUnits(value *float64, units *string) *float64 {
	if value != nil && units != nil {
		switch *units {
		case MgdL, Mgdl:
			intValue := int(*value/MmolLToMgdLConversionFactor*MmolLToMgdLPrecisionFactor + 0.5)
			floatValue := float64(intValue) / MmolLToMgdLPrecisionFactor
			return &floatValue
		}
	}
	return value
}
