package glucose

import (
	"fmt"
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
		}
	}
	return value
}

const (
	// MgdL_Per_MmolL renames MmolLToMgdLConversionFactor in an alternate way.
	MgdL_Per_MmolL = MmolLToMgdLConversionFactor
	// MmolL_Per_MgdL inverts the MmolL_Per_MgdL conversion factor.
	MmolL_Per_MgdL = 1 / MgdL_Per_MmolL
)

// Convert between common blood glucose measurement units.
//
// Results are rounded to five decimal places. However, if fromUnits and
// toUnits are the same, the value is returned unchanged. If either fromUnits
// or toUnits aren't handled, Convert panics.
//
// Convert is modeled to return results as similar as possible to
// NormalizeValueForUnits, but without the hassle of pointer values, and with
// the ability to convert between arbitrary units (as supported).
func Convert(value float64, fromUnits, toUnits string) float64 {
	from, to := normalizeUnitsCase(fromUnits), normalizeUnitsCase(toUnits)
	if from == to {
		return value
	}
	if from == MgdL && to == MmolL {
		return Round5(value * MmolL_Per_MgdL)
	}
	if from == MmolL && to == MgdL {
		return Round5(value * MgdL_Per_MmolL)
	}
	panic(fmt.Errorf("unhandled conversion %q => %q", fromUnits, toUnits))
}

// Round5 rounds a value to 5 decimal places.
func Round5(value float64) float64 {
	return math.Round(value*1e5) / 1e5
}

// normalizeUnitsCase collapses the cAsE of equivalent units into one form.
func normalizeUnitsCase(units string) string {
	switch units {
	case MmolL, Mmoll:
		return MmolL
	case MgdL, Mgdl:
		return MgdL
	}
	return units
}
