package test

import (
	"time"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/test"
)

func RandomMomentFromRange(minimum time.Time, maximum time.Time) *dexcom.Moment {
	return RandomMomentFromTime(test.RandomTimeFromRange(minimum, maximum))
}

func RandomMomentFromTime(tm time.Time) *dexcom.Moment {
	datum := dexcom.NewMoment()
	datum.SystemTime = dexcom.TimeFromRaw(tm.Truncate(time.Second).UTC())
	datum.DisplayTime = dexcom.TimeFromRaw(timeInRandomLocation(datum.SystemTime.Time))
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

func NewObjectFromMoment(datum *dexcom.Moment, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.SystemTime != nil {
		object["systemTime"] = test.NewObjectFromString(datum.SystemTime.String(), objectFormat)
	}
	if datum.DisplayTime != nil {
		object["displayTime"] = test.NewObjectFromString(datum.DisplayTime.String(), objectFormat)
	}
	return object
}

func timeInRandomLocation(tm time.Time) time.Time {
	var location *time.Location
	if offset := test.RandomIntFromRange(-12, 14) * 60 * 60; offset == 0 {
		location = time.UTC
	} else if _, localOffset := tm.Local().Zone(); offset == localOffset {
		location = time.Local
	} else {
		location = time.FixedZone("", offset)
	}
	return tm.In(location)
}
