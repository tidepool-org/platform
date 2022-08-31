package test

import (
	"math"
	"math/rand"
)

func MustInt64(value int64, err error) int64 {
	if err != nil {
		panic(err)
	}
	return value
}

func RandomInt64() int64 {
	return RandomInt64FromRange(RandomInt64Minimum(), RandomInt64Maximum())
}

func RandomInt64FromArray(array []int64) int64 {
	if len(array) == 0 {
		panic("RandomInt64FromArray: array is empty")
	}
	return array[rand.Intn(len(array))]
}

func RandomInt64FromRange(minimum int64, maximum int64) int64 {
	if maximum < minimum {
		panic("RandomInt64FromRange: maximum is not greater than or equal to minimum")
	}
	if minimum < RandomInt64Minimum() {
		minimum = RandomInt64Minimum()
	}
	if maximum > RandomInt64Maximum() {
		maximum = RandomInt64Maximum()
	}
	return minimum + rand.Int63n(maximum-minimum+1)
}

func RandomInt64Maximum() int64 {
	return math.MaxInt64
}

func RandomInt64Minimum() int64 {
	return math.MinInt64
}

func NewObjectFromInt64(value int64, objectFormat ObjectFormat) interface{} {
	switch objectFormat {
	case ObjectFormatJSON:
		return float64(value)
	}
	return value
}
