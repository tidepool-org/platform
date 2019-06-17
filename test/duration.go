package test

import (
	"math/rand"
	"time"
)

func MustDuration(value time.Duration, err error) time.Duration {
	if err != nil {
		panic(err)
	}
	return value
}

func RandomDuration() time.Duration {
	return RandomDurationFromRange(RandomDurationMinimum(), RandomDurationMaximum())
}

func RandomDurationFromArray(array []time.Duration) time.Duration {
	if len(array) == 0 {
		panic("RandomDurationFromArray: array is empty")
	}
	return array[rand.Intn(len(array))]
}

func RandomDurationFromRange(minimum time.Duration, maximum time.Duration) time.Duration {
	if maximum < minimum {
		panic("RandomDurationFromRange: maximum is not greater than or equal to minimum")
	}
	if minimum < RandomDurationMinimum() {
		minimum = RandomDurationMinimum()
	}
	if maximum > RandomDurationMaximum() {
		maximum = RandomDurationMaximum()
	}
	return minimum + time.Duration(rand.Int63n(int64(maximum-minimum+1)))
}

func RandomDurationMaximum() time.Duration {
	return 10 * 365 * 24 * time.Hour
}

func RandomDurationMinimum() time.Duration {
	return -10 * 365 * 24 * time.Hour
}

func NewObjectFromDuration(value time.Duration, objectFormat ObjectFormat) interface{} {
	return value
}
