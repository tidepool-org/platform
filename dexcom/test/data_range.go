package test

import (
	"time"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDataRangeResponse() *dexcom.DataRangeResponse {
	datum := dexcom.NewDataRangeResponse()
	datum.RecordType = pointer.FromString(dexcom.DataRangeResponseRecordType)
	datum.RecordVersion = pointer.FromString(dexcom.DataRangeResponseRecordVersion)
	datum.UserID = pointer.FromString(test.RandomString())
	datum.Calibrations = RandomDataRange()
	datum.EGVs = RandomDataRange()
	datum.Events = RandomDataRange()
	return datum
}

func CloneDataRangeResponse(datum *dexcom.DataRangeResponse) *dexcom.DataRangeResponse {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewDataRangeResponse()
	clone.RecordType = pointer.CloneString(datum.RecordType)
	clone.RecordVersion = pointer.CloneString(datum.RecordVersion)
	clone.UserID = pointer.CloneString(datum.UserID)
	clone.Calibrations = CloneDataRange(datum.Calibrations)
	clone.EGVs = CloneDataRange(datum.EGVs)
	clone.Events = CloneDataRange(datum.Events)
	return clone
}

func RandomDataRange() *dexcom.DataRange {
	datum := dexcom.NewDataRange()
	datum.Start = RandomMomentFromRange(test.RandomTimeMinimum(), time.Now())
	datum.End = RandomMomentFromRange(datum.Start.SystemTime.Time, time.Now())
	return datum
}

func CloneDataRange(datum *dexcom.DataRange) *dexcom.DataRange {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewDataRange()
	clone.Start = CloneMoment(datum.Start)
	clone.End = CloneMoment(datum.End)
	return clone
}
