package test

import (
	"fmt"
	"time"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/test"
)

func RandomMomentFromRange(minimum time.Time, maximum time.Time) *dexcom.Moment {
	datum := dexcom.NewMoment()
	datum.SystemTime = dexcom.TimeFromRaw(test.RandomTimeFromRange(minimum, maximum).Truncate(time.Second).UTC())
	datum.DisplayTime = dexcom.TimeFromRaw(datum.SystemTime.Time.In(randomLocation()))
	return datum
}

func CloneMoment(datum *dexcom.Moment) *dexcom.Moment {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewMoment()
	clone.SystemTime = CloneTime(datum.SystemTime)
	clone.DisplayTime = CloneTime(datum.DisplayTime)
	return clone
}

func randomLocation() *time.Location {
	offsetHours := test.RandomIntFromRange(-12, 14)
	return time.FixedZone(fmt.Sprintf("UTC%+d", offsetHours), offsetHours*60*60)
}
