package test

import (
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/data"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	timeZoneTest "github.com/tidepool-org/platform/time/zone/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

func RandomClockDriftOffset() int {
	return -86400000 + rand.Intn(86400000+86400000)
}

func RandomConversionOffset() int {
	return -9999999999 + rand.Intn(9999999999+9999999999)
}

func RandomVersionInternal() int {
	return rand.Intn(10)
}

func RandomSetID() string {
	return data.NewSetID()
}

func RandomSetIDs() []string {
	return test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(1, 3, RandomSetID)
}

func RandomDataSetClient() *data.DataSetClient {
	datum := data.NewDataSetClient()
	datum.Name = pointer.FromString(netTest.RandomReverseDomain())
	datum.Version = pointer.FromString(netTest.RandomSemanticVersion())
	datum.Private = metadataTest.RandomMetadataMap()
	return datum
}

func CloneDataSetClient(datum *data.DataSetClient) *data.DataSetClient {
	if datum == nil {
		return nil
	}
	clone := data.NewDataSetClient()
	clone.Name = pointer.CloneString(datum.Name)
	clone.Version = pointer.CloneString(datum.Version)
	clone.Private = metadataTest.CloneMetadataMap(datum.Private)
	return clone
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

func RandomDataSet() *data.DataSet {
	createdTime := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now().Add(-30*24*time.Hour))
	modifiedTime := test.RandomTimeFromRange(createdTime, time.Now().Add(-24*time.Hour))
	deletedTime := test.RandomTimeFromRange(modifiedTime, time.Now())

	datum := data.NewDataSet()
	datum.Active = false
	datum.Annotations = metadataTest.RandomMetadataArray()
	datum.ByUser = pointer.FromString(userTest.RandomID())
	datum.Client = RandomDataSetClient()
	datum.ClockDriftOffset = pointer.FromInt(RandomClockDriftOffset())
	datum.ComputerTime = pointer.FromString(test.RandomTime().Format("2006-01-02T15:04:05"))
	datum.ConversionOffset = pointer.FromInt(RandomConversionOffset())
	datum.CreatedTime = pointer.FromTime(createdTime)
	datum.CreatedUserID = pointer.FromString(userTest.RandomID())
	datum.DataSetType = pointer.FromString(test.RandomStringFromArray(data.DataSetTypes()))
	datum.DataState = pointer.FromString(test.RandomStringFromArray(data.DataSetStates()))
	datum.Deduplicator = RandomDeduplicatorDescriptor()
	datum.DeletedTime = pointer.FromTime(deletedTime)
	datum.DeletedUserID = pointer.FromString(userTest.RandomID())
	datum.DeviceID = pointer.FromString(NewDeviceID())
	datum.DeviceManufacturers = pointer.FromStringArray([]string{test.RandomStringFromRange(1, 16), test.RandomStringFromRange(1, 16)})
	datum.DeviceModel = pointer.FromString(test.RandomStringFromRange(1, 32))
	datum.DeviceSerialNumber = pointer.FromString(test.RandomStringFromRange(1, 16))
	datum.DeviceTags = pointer.FromStringArray(test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(data.DeviceTags()), data.DeviceTags()))
	datum.DeviceTime = pointer.FromString(test.RandomTime().Format("2006-01-02T15:04:05"))
	datum.ID = pointer.FromString(RandomID())
	datum.ModifiedTime = pointer.FromTime(modifiedTime)
	datum.ModifiedUserID = pointer.FromString(userTest.RandomID())
	datum.Payload = metadataTest.RandomMetadata()
	datum.Provenance = RandomProvenance()
	datum.State = pointer.FromString(test.RandomStringFromArray(data.DataSetStates()))
	datum.Time = pointer.FromTime(test.RandomTime())
	datum.TimeProcessing = pointer.FromString(data.TimeProcessingUTCBootstrapping)
	datum.TimeZoneName = pointer.FromString(timeZoneTest.RandomName())
	datum.TimeZoneOffset = pointer.FromInt(RandomTimeZoneOffset())
	datum.UploadID = pointer.FromString(RandomSetID())
	datum.UserID = pointer.FromString(userTest.RandomID())
	datum.Version = pointer.FromString(netTest.RandomSemanticVersion())
	datum.VersionInternal = RandomVersionInternal()
	return datum
}

func CloneDataSet(datum *data.DataSet) *data.DataSet {
	if datum == nil {
		return nil
	}
	clone := data.NewDataSet()
	clone.Active = datum.Active
	clone.Annotations = metadataTest.CloneMetadataArray(datum.Annotations)
	clone.ByUser = pointer.CloneString(datum.ByUser)
	clone.Client = CloneDataSetClient(datum.Client)
	clone.ClockDriftOffset = pointer.CloneInt(datum.ClockDriftOffset)
	clone.ComputerTime = pointer.CloneString(datum.ComputerTime)
	clone.ConversionOffset = pointer.CloneInt(datum.ConversionOffset)
	clone.CreatedTime = pointer.CloneTime(datum.CreatedTime)
	clone.CreatedUserID = pointer.CloneString(datum.CreatedUserID)
	clone.DataSetType = pointer.CloneString(datum.DataSetType)
	clone.DataState = pointer.CloneString(datum.DataState)
	clone.Deduplicator = CloneDeduplicatorDescriptor(datum.Deduplicator)
	clone.DeletedTime = pointer.CloneTime(datum.DeletedTime)
	clone.DeletedUserID = pointer.CloneString(datum.DeletedUserID)
	clone.DeviceID = pointer.CloneString(datum.DeviceID)
	clone.DeviceManufacturers = pointer.CloneStringArray(datum.DeviceManufacturers)
	clone.DeviceModel = pointer.CloneString(datum.DeviceModel)
	clone.DeviceSerialNumber = pointer.CloneString(datum.DeviceSerialNumber)
	clone.DeviceTags = pointer.CloneStringArray(datum.DeviceTags)
	clone.DeviceTime = pointer.CloneString(datum.DeviceTime)
	clone.ID = pointer.CloneString(datum.ID)
	clone.ModifiedTime = pointer.CloneTime(datum.ModifiedTime)
	clone.ModifiedUserID = pointer.CloneString(datum.ModifiedUserID)
	clone.Payload = metadataTest.CloneMetadata(datum.Payload)
	clone.Provenance = CloneProvenance(datum.Provenance)
	clone.State = pointer.CloneString(datum.State)
	clone.Time = pointer.CloneTime(datum.Time)
	clone.TimeProcessing = pointer.CloneString(datum.TimeProcessing)
	clone.TimeZoneName = pointer.CloneString(datum.TimeZoneName)
	clone.TimeZoneOffset = pointer.CloneInt(datum.TimeZoneOffset)
	clone.UploadID = pointer.CloneString(datum.UploadID)
	clone.UserID = pointer.CloneString(datum.UserID)
	clone.Version = pointer.CloneString(datum.Version)
	clone.VersionInternal = datum.VersionInternal
	return clone
}

func RandomTimeZoneOffset() int {
	return -4440 + rand.Intn(4440+6960)
}
