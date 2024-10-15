package test

import (
	"time"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDataRangesResponse() *dexcom.DataRangesResponse {
	datum := dexcom.NewDataRangesResponse()
	datum.RecordType = pointer.FromString(dexcom.DataRangesResponseRecordType)
	datum.RecordVersion = pointer.FromString(dexcom.DataRangesResponseRecordVersion)
	datum.UserID = pointer.FromString(test.RandomString())
	datum.Calibrations = RandomDataRange()
	datum.EGVs = RandomDataRange()
	datum.Events = RandomDataRange()
	return datum
}

func CloneDataRangesResponse(datum *dexcom.DataRangesResponse) *dexcom.DataRangesResponse {
	if datum == nil {
		return nil
	}
	clone := dexcom.NewDataRangesResponse()
	clone.RecordType = pointer.CloneString(datum.RecordType)
	clone.RecordVersion = pointer.CloneString(datum.RecordVersion)
	clone.UserID = pointer.CloneString(datum.UserID)
	clone.Calibrations = CloneDataRange(datum.Calibrations)
	clone.EGVs = CloneDataRange(datum.EGVs)
	clone.Events = CloneDataRange(datum.Events)
	return clone
}

func NewObjectFromDataRangesResponse(datum *dexcom.DataRangesResponse, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.RecordType != nil {
		object["recordType"] = test.NewObjectFromString(*datum.RecordType, objectFormat)
	}
	if datum.RecordVersion != nil {
		object["recordVersion"] = test.NewObjectFromString(*datum.RecordVersion, objectFormat)
	}
	if datum.UserID != nil {
		object["userId"] = test.NewObjectFromString(*datum.UserID, objectFormat)
	}
	if datum.Calibrations != nil {
		object["calibrations"] = NewObjectFromDataRange(datum.Calibrations, objectFormat)
	}
	if datum.EGVs != nil {
		object["egvs"] = NewObjectFromDataRange(datum.EGVs, objectFormat)
	}
	if datum.Events != nil {
		object["events"] = NewObjectFromDataRange(datum.Events, objectFormat)
	}
	return object
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

func NewObjectFromDataRange(datum *dexcom.DataRange, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Start != nil {
		object["start"] = NewObjectFromMoment(datum.Start, objectFormat)
	}
	if datum.End != nil {
		object["end"] = NewObjectFromMoment(datum.End, objectFormat)
	}
	return object
}
