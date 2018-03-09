package test

import (
	"math"
	"math/rand"
)

func RandomBool() bool {
	return rand.Intn(2) == 0
}

func RandomFloat64FromArray(array []float64) float64 {
	return array[rand.Intn(len(array))]
}

func RandomFloat64FromRange(minimum float64, maximum float64) float64 {
	if minimum != -math.MaxFloat64 || maximum != math.MaxFloat64 {
		return minimum + (maximum-minimum)*rand.Float64()
	}
	return rand.NormFloat64()
}

func RandomIntFromArray(array []int) int {
	return array[rand.Intn(len(array))]
}

func RandomIntFromRange(minimum int, maximum int) int {
	return minimum + rand.Intn(maximum-minimum+1)
}

func RandomStringFromArray(array []string) string {
	return array[rand.Intn(len(array))]
}

func RandomStringArrayFromArray(minimumLength int, maximumLength int, duplicates bool, array []string) []string {
	resultLength := minimumLength + rand.Intn(maximumLength-minimumLength+1)
	result := make([]string, resultLength)
	if duplicates {
		for resultIndex := range result {
			result[resultIndex] = RandomStringFromArray(array)
		}
	} else {
		resultIndex := 0
		arrayLength := len(array)
		for arrayIndex, arrayValue := range array {
			if rand.Float64() < float64(resultLength-resultIndex)/float64(arrayLength-arrayIndex) {
				result[resultIndex] = arrayValue
				resultIndex++
			}
		}
	}
	return result
}
