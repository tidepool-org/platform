package devices

import (
	"math/big"
	"sort"
	"strconv"

	"github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var scale = big.NewInt(1000000000)

func ValidateBasalRateSchedule(basalRateSchedule pump.BasalRateStartArray, guardRail *api.BasalRatesGuardRail, validator structure.Validator) {
	validValues := generateAllValidValues(guardRail.AbsoluteBounds)
	for i, basalRate := range basalRateSchedule {
		ValidateIncrementIfValueNotNil(basalRate.Rate, validValues, validator.WithReference(strconv.Itoa(i)).WithReference("rate"))
	}
}

func ValidateBloodGlucoseTargetSchedule(bloodGlucoseTargetSchedule pump.BloodGlucoseTargetStartArray, guardRail *api.CorrectionRangeGuardRail, validator structure.Validator) {
	validValues := generateValidValues(guardRail.AbsoluteBounds)
	for i, bloodGlucoseTarget := range bloodGlucoseTargetSchedule {
		ValidateIncrementIfValueNotNil(bloodGlucoseTarget.Target.High, validValues, validator.WithReference(strconv.Itoa(i)).WithReference("high"))
		ValidateIncrementIfValueNotNil(bloodGlucoseTarget.Target.Low, validValues, validator.WithReference(strconv.Itoa(i)).WithReference("low"))
		ValidateIncrementIfValueNotNil(bloodGlucoseTarget.Target.Range, validValues, validator.WithReference(strconv.Itoa(i)).WithReference("range"))
		ValidateIncrementIfValueNotNil(bloodGlucoseTarget.Target.Target, validValues, validator.WithReference(strconv.Itoa(i)).WithReference("target"))
	}
}

func ValidateCarbohydrateRatioSchedule(carbohydrateRatioSchedule pump.CarbohydrateRatioStartArray, guardRail *api.CarbohydrateRatioGuardRail, validator structure.Validator) {
	validValues := generateValidValues(guardRail.AbsoluteBounds)
	for i, carbRatio := range carbohydrateRatioSchedule {
		ValidateIncrementIfValueNotNil(carbRatio.Amount, validValues, validator.WithReference(strconv.Itoa(i)).WithReference("amount"))
	}
}

func ValidateInsulinSensitivitySchedule(insulinSensitivitySchedule pump.InsulinSensitivityStartArray, guardRail *api.InsulinSensitivityGuardRail, validator structure.Validator) {
	validValues := generateValidValues(guardRail.AbsoluteBounds)
	for i, insulinSensitivity := range insulinSensitivitySchedule {
		ValidateIncrementIfValueNotNil(insulinSensitivity.Amount, validValues, validator.WithReference(strconv.Itoa(i)).WithReference("amount"))
	}
}

func ValidateBasalRateMaximum(basalRateMaximum pump.BasalRateMaximum, guardRail *api.BasalRateMaximumGuardRail, validator structure.Validator) {
	validValues := generateValidValues(guardRail.AbsoluteBounds)
	ValidateIncrementIfValueNotNil(basalRateMaximum.Value, validValues, validator.WithReference("value"))
}

func ValidateBolusAmountMaximum(bolusAmountMaximum pump.BolusAmountMaximum, guardRail *api.BolusAmountMaximumGuardRail, validator structure.Validator) {
	validValues := generateValidValues(guardRail.AbsoluteBounds)
	ValidateIncrementIfValueNotNil(bolusAmountMaximum.Value, validValues, validator.WithReference("value"))
}

func ValidateIncrementIfValueNotNil(value *float64, validValues []float64, validator structure.Validator) {
	if value != nil && !IsValidIncrement(*value, validValues) {
		validator.ReportError(structureValidator.ErrorValueNotValid())
	}
}

func IsValidIncrement(value float64, validValues []float64) bool {
	i := positionInSortedArray(value, validValues)
	return i < len(validValues) && validValues[i] == value
}

func positionInSortedArray(value float64, validValues []float64) int {
	return sort.Search(len(validValues), func(i int) bool { return validValues[i] >= value })
}

func generateAllValidValues(bounds []*api.AbsoluteBounds) []float64 {
	validValues := make([]float64, 0)
	for _, b := range bounds {
		for _, value := range generateValidValues(b) {
			// insert in a sorted array
			i := positionInSortedArray(value, validValues)
			if i >= len(validValues) || validValues[i] != value {
				validValues = append(validValues, value)
				copy(validValues[i+1:], validValues[i:])
				validValues[i] = value
			}
		}
	}
	return validValues
}

func generateValidValues(bounds *api.AbsoluteBounds) []float64 {
	valid := make([]float64, 0)
	if bounds == nil {
		return valid
	}

	increment := fixedDecimalToBigInt(bounds.Increment)
	minimum := fixedDecimalToBigInt(bounds.Minimum)
	maximum := fixedDecimalToBigInt(bounds.Maximum)

	current := big.NewInt(minimum.Int64())
	for current.Cmp(maximum) < 0 {
		// calculate the float64 approximation of the current increment
		value, _ := bigIntToFloat(current).Float64()
		valid = append(valid, value)

		current.Add(current, increment)
	}
	return valid
}

func fixedDecimalToBigInt(decimal *api.FixedDecimal) *big.Int {
	bi := big.NewInt(0)
	bi.Mul(big.NewInt(int64(decimal.Units)), scale).Add(bi, big.NewInt(int64(decimal.Nanos)))
	return bi
}

func bigIntToFloat(value *big.Int) *big.Float {
	bf := big.NewFloat(0).SetInt(value)
	bf.Quo(bf, big.NewFloat(0).SetInt(scale))
	return bf
}
