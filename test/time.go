package test

import (
	"math/rand"
	"time"
)

const TimeLimit = 365 * 24 * time.Hour

func TimeMaximum() time.Time {
	return time.Now().Add(TimeLimit)
}

func TimeMinimum() time.Time {
	return time.Now().Add(-TimeLimit)
}

func NewTime() time.Time {
	return NewTimeInRange(TimeMinimum(), TimeMaximum())
}

func NewTimeInRange(earliest time.Time, latest time.Time) time.Time {
	return earliest.Add(time.Duration(rand.Int63n(int64(latest.Sub(earliest)))))
}
