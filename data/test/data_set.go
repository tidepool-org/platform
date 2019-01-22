package test

import (
	"math/rand"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	timeZoneTest "github.com/tidepool-org/platform/time/zone/test"
)

func RandomSetID() string {
	return data.NewSetID()
}

func RandomSetIDs() []string {
	return test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(1, 3, RandomSetID)
}

func RandomDataSetUpdate() *data.DataSetUpdate {
	datum := data.NewDataSetUpdate()
	datum.Active = pointer.FromBool(false)
	datum.DeviceID = pointer.FromString(NewDeviceID())
	datum.DeviceModel = pointer.FromString(test.RandomStringFromRange(1, 32))
	datum.DeviceSerialNumber = pointer.FromString(test.RandomStringFromRange(1, 16))
	datum.Deduplicator = RandomDeduplicatorDescriptor()
	datum.State = pointer.FromString(test.RandomStringFromArray([]string{"closed", "open"}))
	datum.Time = pointer.FromTime(test.RandomTime())
	datum.TimeZoneName = pointer.FromString(timeZoneTest.RandomName())
	datum.TimeZoneOffset = pointer.FromInt(RandomTimeZoneOffset())
	return datum
}

func RandomTimeZoneOffset() int {
	return -4440 + rand.Intn(4440+6960)
}
