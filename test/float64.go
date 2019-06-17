package test

import (
	"math"
	"math/rand"
)

func MustFloat64(value float64, err error) float64 {
	if err != nil {
		panic(err)
	}
	return value
}

func RandomFloat64() float64 {
	return RandomFloat64FromRange(RandomFloat64Minimum(), RandomFloat64Maximum())
}

func RandomFloat64FromArray(array []float64) float64 {
	if len(array) == 0 {
		panic("RandomFloat64FromArray: array is empty")
	}
	return array[rand.Intn(len(array))]
}

func RandomFloat64FromRange(minimum float64, maximum float64) float64 {
	if maximum < minimum {
		panic("RandomFloat64FromRange: maximum is not greater than or equal to minimum")
	}
	if minimum < RandomFloat64Minimum() {
		minimum = RandomFloat64Minimum()
	}
	if maximum > RandomFloat64Maximum() {
		maximum = RandomFloat64Maximum()
	}
	return minimum + (maximum-minimum+math.SmallestNonzeroFloat64)*rand.Float64()
}

func RandomFloat64Maximum() float64 {
	return math.MaxFloat32
}

func RandomFloat64Minimum() float64 {
	return -math.MaxFloat32
}

func NewObjectFromFloat64(value float64, objectFormat ObjectFormat) interface{} {
	return value
}
