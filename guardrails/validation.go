package guardrails

import (
	"math/big"
	"sort"

	devices "github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var scale = big.NewInt(1000000000)

func ValidateValueIfNotNil(value *float64, validValues []float64, validator structure.Validator) {
	if value != nil && !IsValidValue(*value, validValues) {
		validator.ReportError(structureValidator.ErrorValueNotValid())
	}
}

func IsValidValue(value float64, validValues []float64) bool {
	i := positionInSortedArray(value, validValues)
	return i < len(validValues) && validValues[i] == value
}

func positionInSortedArray(value float64, validValues []float64) int {
	return sort.Search(len(validValues), func(i int) bool { return validValues[i] >= value })
}

func generateAllValidValues(bounds []*devices.AbsoluteBounds) []float64 {
	validValues := make([]float64, 0)
	for _, b := range bounds {
		for _, value := range generateValidValuesFromAbsoluteBounds(b) {
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

func generateValidValuesFromAbsoluteBounds(bounds *devices.AbsoluteBounds) []float64 {
	valid := make([]float64, 0)
	if bounds == nil {
		return valid
	}

	increment := fixedDecimalToBigInt(bounds.Increment)
	minimum := fixedDecimalToBigInt(bounds.Minimum)
	maximum := fixedDecimalToBigInt(bounds.Maximum)

	current := big.NewInt(minimum.Int64())
	for current.Cmp(maximum) <= 0 {
		// calculate the float64 approximation of the current increment
		value, _ := bigIntToFloat(current).Float64()
		valid = append(valid, value)

		current.Add(current, increment)
	}
	return valid
}

func discardValuesSmallerThan(values []float64, minimum float64) []float64 {
	valid := make([]float64, 0)
	for _, v := range values {
		if v >= minimum {
			valid = append(valid, v)
		}
	}

	return valid
}

func discardValuesLargerThan(values []float64, maximum float64) []float64 {
	valid := make([]float64, 0)
	for _, v := range values {
		if v >= maximum {
			valid = append(valid, v)
		}
	}

	return valid
}

func fixedDecimalToBigInt(decimal *devices.FixedDecimal) *big.Int {
	bi := big.NewInt(0)
	bi.Mul(big.NewInt(int64(decimal.Units)), scale).Add(bi, big.NewInt(int64(decimal.Nanos)))
	return bi
}

func bigIntToFloat(value *big.Int) *big.Float {
	bf := big.NewFloat(0).SetInt(value)
	bf.Quo(bf, big.NewFloat(0).SetInt(scale))
	return bf
}
