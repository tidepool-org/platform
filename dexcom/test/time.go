package test

import (
	"time"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/test"
)

func RandomSystemTime() *dexcom.Time {
	datum := dexcom.NewTime()
	datum.Time = test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).UTC().Truncate(time.Second)
	return datum
}

func RandomDisplayTime() *dexcom.Time {
	datum := dexcom.NewTime()
	datum.Time = test.RandomTime().UTC().Truncate(time.Second)
	return datum
}

func RandomTime() *dexcom.Time {
	datum := dexcom.NewTime()
	datum.Time = test.RandomTime().UTC().Truncate(time.Second)
	return datum
}

func CloneTime(datum *dexcom.Time) *dexcom.Time {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewTime()
	clone.Time = datum.Time
	return clone
}
