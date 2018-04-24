package test

import (
	"math"
	"math/rand"
)

func RandomIntFromRange(minimum int, maximum int) int {
	return minimum + rand.Intn(maximum-minimum+1)
}

func RandomFloat64FromRange(minimum float64, maximum float64) float64 {
	if minimum != -math.MaxFloat64 || maximum != math.MaxFloat64 {
		return minimum + (maximum-minimum)*rand.Float64()
	}
	return rand.NormFloat64()
}

func RandomStringFromStringArray(strs []string) string {
	return strs[rand.Intn(len(strs))]
}

func RandomStringArrayFromStringArray(minimumLength int, maximumLength int, duplicates bool, strs []string) []string {
	resultLength := minimumLength + rand.Intn(maximumLength-minimumLength+1)
	result := make([]string, resultLength)
	if duplicates {
		for resultIndex := range result {
			result[resultIndex] = RandomStringFromStringArray(strs)
		}
	} else {
		resultIndex := 0
		strsLength := len(strs)
		for strsIndex, strsValue := range strs {
			if rand.Float64() < float64(resultLength-resultIndex)/float64(strsLength-strsIndex) {
				result[resultIndex] = strsValue
				resultIndex++
			}
		}
	}
	return result
}
