package test

import (
	"math"
	"math/rand"
)

func MustInt(value int, err error) int {
	if err != nil {
		panic(err)
	}
	return value
}

func RandomInt() int {
	return RandomIntFromRange(RandomIntMinimum(), RandomIntMaximum())
}

func RandomIntFromArray(array []int) int {
	if len(array) == 0 {
		panic("RandomIntFromArray: array is empty")
	}
	return array[rand.Intn(len(array))]
}

func RandomIntFromRange(minimum int, maximum int) int {
	if maximum < minimum {
		panic("RandomIntFromRange: maximum is not greater than or equal to minimum")
	}
	if minimum < RandomIntMinimum() {
		minimum = RandomIntMinimum()
	}
	if maximum > RandomIntMaximum() {
		maximum = RandomIntMaximum()
	}
	return minimum + rand.Intn(maximum-minimum+1)
}

func RandomIntMaximum() int {
	return math.MaxInt32
}

func RandomIntMinimum() int {
	return math.MinInt32
}

func NewObjectFromInt(value int, objectFormat ObjectFormat) interface{} {
	switch objectFormat {
	case ObjectFormatJSON:
		return float64(value)
	case ObjectFormatBSON:
		if value < math.MinInt32 || value > math.MaxInt32 {
			return int64(value)
		}
	}
	return value
}
