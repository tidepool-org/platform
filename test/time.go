package test

import (
	"math/rand"
	"time"

	"github.com/onsi/gomega"
	gomegaGstruct "github.com/onsi/gomega/gstruct"
	gomegaTypes "github.com/onsi/gomega/types"
)

func PastFarTime() time.Time {
	return now.AddDate(-30, 0, 0)
}

func PastNearTime() time.Time {
	return now.AddDate(0, -1, 0)
}

func FutureNearTime() time.Time {
	return now.AddDate(0, 1, 0)
}

func FutureFarTime() time.Time {
	return now.AddDate(30, 0, 0)
}

func MustTime(value time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return value
}

func RandomTimeBeforeNow() time.Time {
	return RandomTimeBefore(time.Now())
}

func RandomTimeAfterNow() time.Time {
	return RandomTimeAfter(time.Now())
}

func RandomTimeBefore(value time.Time) time.Time {
	return RandomTimeFromRange(RandomTimeMinimum(), PinnedTime(value))
}

func RandomTimeAfter(value time.Time) time.Time {
	return RandomTimeFromRange(PinnedTime(value), RandomTimeMaximum())
}

func RandomTime() time.Time {
	return RandomTimeFromRange(RandomTimeMinimum(), RandomTimeMaximum())
}

func RandomTimeFromArray(array []time.Time) time.Time {
	if len(array) == 0 {
		panic("RandomTimeFromArray: array is empty")
	}
	return array[rand.Intn(len(array))]
}

func RandomTimeFromRange(minimum time.Time, maximum time.Time) time.Time {
	if maximum.Before(minimum) {
		panic("RandomTimeFromRange: maximum is not greater than or equal to minimum")
	}
	minimum = PinnedTime(minimum)
	maximum = PinnedTime(maximum)
	if duration := maximum.Sub(minimum); duration != 0 {
		return minimum.Add(time.Duration(rand.Int63n(int64(duration)))).Truncate(time.Millisecond)
	} else {
		return minimum.Truncate(time.Millisecond)
	}
}

func RandomTimeMaximum() time.Time {
	return now.Add(RandomDurationMaximum()).Truncate(time.Millisecond)
}

func RandomTimeMinimum() time.Time {
	return now.Add(RandomDurationMinimum()).Truncate(time.Millisecond)
}

func NewObjectFromTime(value time.Time, objectFormat ObjectFormat) interface{} {
	switch objectFormat {
	case ObjectFormatJSON:
		return value.Format(time.RFC3339Nano)
	}
	return value
}

func PinnedTime(value time.Time) time.Time {
	if value.Before(RandomTimeMinimum()) {
		return RandomTimeMinimum()
	} else if value.After(RandomTimeMaximum()) {
		return RandomTimeMaximum()
	} else {
		return value.Truncate(time.Millisecond)
	}
}

func MatchTime(datum *time.Time) gomegaTypes.GomegaMatcher {
	if datum == nil {
		return gomega.BeNil()
	}
	return gomegaGstruct.PointTo(gomega.BeTemporally("==", *datum))
}

var now = time.Now().Truncate(time.Millisecond).UTC()
