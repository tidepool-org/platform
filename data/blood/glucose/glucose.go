package glucose

import (
	"math"

	"github.com/tidepool-org/platform/pointer"
)

const (
	MmolL = "mmol/L"
	Mmoll = "mmol/l"

	MgdL = "mg/dL"
	Mgdl = "mg/dl"

	MmolLMinimum float64 = 0.0
	MmolLMaximum float64 = 55.0

	MgdLMinimum float64 = 0.0
	MgdLMaximum float64 = 1000.0

	// MmolLToMgdLConversionFactor is MgdL Per MmolL.
	//
	// Reminder: The molecular mass of glucose is ≈ 180 g/mol.
	//
	// MmolLToMgdLConversionFactor can be used like this:
	//   140 MgdL / MmolLToMgdLConversionFactor = 7.77105 mmol/L
	//   7.77105 mmol/L * MmolLToMgdLConversionFactor = 140.000 mg/dL
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
			return MmolLMinimum, MmolLMaximum
		case MgdL, Mgdl:
			return MgdLMinimum, MgdLMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

func NormalizeUnits(units *string) *string {
	if units != nil {
		switch *units {
		case MmolL, Mmoll, MgdL, Mgdl:
			return pointer.FromString(MmolL)
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

		case MmolL, Mmoll:

			mgdlVal := *value * MmolLToMgdLConversionFactor

			intValue := int(mgdlVal/MmolLToMgdLConversionFactor*MmolLToMgdLPrecisionFactor + 0.5)
			floatValue := float64(intValue) / MmolLToMgdLPrecisionFactor
			return &floatValue

			//if len([]rune(strVal)) > 8 {
			// floatValue := math.Floor(*value*MmolLToMgdLPrecisionFactor) / MmolLToMgdLPrecisionFactor
			// return &floatValue
			//}
			//return value
		}
	}
	return value
}
