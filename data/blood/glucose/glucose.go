package glucose

import (
	"math"
	"slices"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
)

const (
	MmolL = "mmol/L"
	Mmoll = "mmol/l"

	MmolLMinute = "mmol/L/minute"

	MgdL = "mg/dL"
	Mgdl = "mg/dl"

	MgdLMinute = "mg/dL/minute"

	MmolLMinimum float64 = 0.0
	MmolLMaximum float64 = 55.0

	MmolLMinuteMinimum float64 = -5.5
	MmolLMinuteMaximum float64 = 5.5

	MgdLMinimum float64 = 0.0
	MgdLMaximum float64 = 1000.0

	MgdLMinuteMinimum float64 = -100.0
	MgdLMinuteMaximum float64 = 100.0

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

func RateUnits() []string {
	return []string{MmolLMinute, MgdLMinute}
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

// NormalizeValueForUnits converts Mg/dL values to Mmol/L.
//
// Values paired with any other units (including nil or typos) are NOT normalized or
// modified, but are returned as-is.
func NormalizeValueForUnits(value *float64, units *string) *float64 {
	if value != nil && units != nil {
		switch *units {
		case MgdL, Mgdl:
			intValue := int(*value/MmolLToMgdLConversionFactor*MmolLToMgdLPrecisionFactor + math.Copysign(0.5, *value))
			floatValue := float64(intValue) / MmolLToMgdLPrecisionFactor
			return &floatValue
		}
	}
	return value
}

// NormalizeValueForUnitsSafer behaves like NormalizeValueForUnits, but with more safety.
//
// Where "safety" means returning an error in certain cases where its namesake would return
// the original value.
//
// Notable cases include:
//   - when units is nil but values is not
//   - when units are not a recognized value in Units()
func NormalizeValueForUnitsSafer(value *float64, units *string) (float64, error) {
	if units == nil {
		return 0, errors.New("unable to normalize: unhandled units: nil")
	}
	if !slices.Contains(Units(), *units) {
		return 0, errors.Newf("unable to normalize: unhandled units: %s", *units)
	}
	if value == nil {
		return 0, errors.New("unable to normalize: unhandled value: nil")
	}
	normalized := NormalizeValueForUnits(value, units)
	if normalized == nil {
		// The only time this should happen is when a nil value was passed, and that case is
		// covered above, but it's better to be safe than sorry.
		return 0, errors.New("unable to normalize: normalization returned nil")
	}
	return *normalized, nil
}

func ValueRangeForRateUnits(rateUnits *string) (float64, float64) {
	if rateUnits != nil {
		switch *rateUnits {
		case MmolLMinute:
			return MmolLMinuteMinimum, MmolLMinuteMaximum
		case MgdLMinute:
			return MgdLMinuteMinimum, MgdLMinuteMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

func NormalizeRateUnits(rateUnits *string) *string {
	if rateUnits != nil {
		switch *rateUnits {
		case MmolLMinute, MgdLMinute:
			return pointer.FromString(MmolLMinute)
		}
	}
	return rateUnits
}

func NormalizeValueForRateUnits(value *float64, rateUnits *string) *float64 {
	if value != nil && rateUnits != nil {
		switch *rateUnits {
		case MgdLMinute:
			intValue := int(*value/MmolLToMgdLConversionFactor*MmolLToMgdLPrecisionFactor + math.Copysign(0.5, *value))
			floatValue := float64(intValue) / MmolLToMgdLPrecisionFactor
			return &floatValue
		}
	}
	return value
}
