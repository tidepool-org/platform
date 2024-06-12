package test

import (
	"time"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/test"
)

func RandomSystemTime() *dexcom.Time {
	return dexcom.TimeFromRaw(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second).UTC())
}

func RandomDisplayTime() *dexcom.Time {
	return dexcom.TimeFromRaw(test.RandomTime().Truncate(time.Second))
}

func RandomTime() *dexcom.Time {
	return dexcom.TimeFromRaw(test.RandomTime().Truncate(time.Second))
}

func CloneTime(datum *dexcom.Time) *dexcom.Time {
	return dexcom.TimeFromTime(datum)
}
