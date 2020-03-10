package test

import (
	"time"

	associationTest "github.com/tidepool-org/platform/association/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	locationTest "github.com/tidepool-org/platform/location/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	originTest "github.com/tidepool-org/platform/origin/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	timeZoneTest "github.com/tidepool-org/platform/time/zone/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

func NewBase() *types.Base {
	createdTime := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now().Add(-30*24*time.Hour))
	archivedTime := test.RandomTimeFromRange(createdTime, time.Now().Add(-7*24*time.Hour))
	modifiedTime := test.RandomTimeFromRange(archivedTime, time.Now().Add(-24*time.Hour))
	deletedTime := test.RandomTimeFromRange(modifiedTime, time.Now())

	datum := &types.Base{}
	datum.Active = false
	datum.Annotations = metadataTest.RandomMetadataArray()
	datum.Associations = associationTest.RandomAssociationArray()
	datum.ArchivedDataSetID = pointer.FromString(dataTest.RandomSetID())
	datum.ArchivedTime = pointer.FromString(archivedTime.Format(time.RFC3339Nano))
	datum.ClockDriftOffset = pointer.FromInt(NewClockDriftOffset())
	datum.ConversionOffset = pointer.FromInt(NewConversionOffset())
	datum.CreatedTime = pointer.FromString(createdTime.Format(time.RFC3339Nano))
	datum.CreatedUserID = pointer.FromString(userTest.RandomID())
	datum.Deduplicator = dataTest.RandomDeduplicatorDescriptor()
	datum.DeletedTime = pointer.FromString(deletedTime.Format(time.RFC3339Nano))
	datum.DeletedUserID = pointer.FromString(userTest.RandomID())
	datum.DeviceID = pointer.FromString(dataTest.NewDeviceID())
	datum.DeviceTime = pointer.FromString(test.RandomTime().Format("2006-01-02T15:04:05"))
	datum.GUID = pointer.FromString(dataTest.RandomID())
	datum.ID = pointer.FromString(dataTest.RandomID())
	datum.Location = locationTest.RandomLocation()
	datum.ModifiedTime = pointer.FromString(modifiedTime.Format(time.RFC3339Nano))
	datum.ModifiedUserID = pointer.FromString(userTest.RandomID())
	datum.Notes = pointer.FromStringArray([]string{NewNote(1, 20), NewNote(1, 20)})
	datum.Origin = originTest.RandomOrigin()
	datum.Payload = metadataTest.RandomMetadata()
	datum.SchemaVersion = 2
	datum.Source = pointer.FromString("carelink")
	datum.Tags = pointer.FromStringArray([]string{NewTag(1, 10)})
	datum.Time = pointer.FromString(test.RandomTime().Format(time.RFC3339Nano))
	datum.TimeZoneName = pointer.FromString(timeZoneTest.RandomName())
	datum.TimeZoneOffset = pointer.FromInt(NewTimeZoneOffset())
	datum.Type = NewType()
	datum.UploadID = pointer.FromString(dataTest.RandomSetID())
	datum.UserID = pointer.FromString(userTest.RandomID())
	datum.VersionInternal = NewVersionInternal()
	return datum
}

func CloneBase(datum *types.Base) *types.Base {
	if datum == nil {
		return nil
	}
	clone := &types.Base{}
	clone.Active = datum.Active
	clone.Annotations = metadataTest.CloneMetadataArray(datum.Annotations)
	clone.Associations = associationTest.CloneAssociationArray(datum.Associations)
	clone.ArchivedDataSetID = pointer.CloneString(datum.ArchivedDataSetID)
	clone.ArchivedTime = pointer.CloneString(datum.ArchivedTime)
	clone.ClockDriftOffset = pointer.CloneInt(datum.ClockDriftOffset)
	clone.ConversionOffset = pointer.CloneInt(datum.ConversionOffset)
	clone.CreatedTime = pointer.CloneString(datum.CreatedTime)
	clone.CreatedUserID = pointer.CloneString(datum.CreatedUserID)
	clone.Deduplicator = dataTest.CloneDeduplicatorDescriptor(datum.Deduplicator)
	clone.DeletedTime = pointer.CloneString(datum.DeletedTime)
	clone.DeletedUserID = pointer.CloneString(datum.DeletedUserID)
	clone.DeviceID = pointer.CloneString(datum.DeviceID)
	clone.DeviceTime = pointer.CloneString(datum.DeviceTime)
	clone.GUID = pointer.CloneString(datum.GUID)
	clone.ID = pointer.CloneString(datum.ID)
	clone.Location = locationTest.CloneLocation(datum.Location)
	clone.ModifiedTime = pointer.CloneString(datum.ModifiedTime)
	clone.ModifiedUserID = pointer.CloneString(datum.ModifiedUserID)
	clone.Notes = pointer.CloneStringArray(datum.Notes)
	clone.Origin = originTest.CloneOrigin(datum.Origin)
	clone.Payload = metadataTest.CloneMetadata(datum.Payload)
	clone.SchemaVersion = datum.SchemaVersion
	clone.Source = pointer.CloneString(datum.Source)
	clone.Tags = pointer.CloneStringArray(datum.Tags)
	clone.Time = pointer.CloneString(datum.Time)
	clone.TimeZoneName = pointer.CloneString(datum.TimeZoneName)
	clone.TimeZoneOffset = pointer.CloneInt(datum.TimeZoneOffset)
	clone.Type = datum.Type
	clone.UploadID = pointer.CloneString(datum.UploadID)
	clone.UserID = pointer.CloneString(datum.UserID)
	clone.VersionInternal = datum.VersionInternal
	return clone
}
