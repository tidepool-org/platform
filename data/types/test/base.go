package test

import (
	"time"

	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	dataTypesCommonAssociationTest "github.com/tidepool-org/platform/data/types/common/association/test"
	dataTypesCommonLocationTest "github.com/tidepool-org/platform/data/types/common/location/test"
	dataTypesCommonOriginTest "github.com/tidepool-org/platform/data/types/common/origin/test"
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
	datum.Annotations = dataTest.NewBlobArray()
	datum.Associations = dataTypesCommonAssociationTest.NewAssociationArray()
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
	datum.Location = dataTypesCommonLocationTest.NewLocation()
	datum.ModifiedTime = pointer.FromString(modifiedTime.Format(time.RFC3339Nano))
	datum.ModifiedUserID = pointer.FromString(userTest.RandomID())
	datum.Notes = pointer.FromStringArray([]string{NewNote(1, 20), NewNote(1, 20)})
	datum.Origin = dataTypesCommonOriginTest.NewOrigin()
	datum.Payload = dataTest.NewBlob()
	datum.SchemaVersion = 2
	datum.Source = pointer.FromString("carelink")
	datum.Tags = pointer.FromStringArray([]string{NewTag(1, 10)})
	datum.Time = pointer.FromString(test.RandomTime().Format(time.RFC3339Nano))
	datum.TimeZoneName = pointer.FromString(timeZoneTest.RandomName())
	datum.TimeZoneOffset = pointer.FromInt(NewTimeZoneOffset())
	datum.Type = NewType()
	datum.UploadID = pointer.FromString(dataTest.RandomSetID())
	datum.UserID = pointer.FromString(userTest.RandomID())
	datum.Version = NewVersion()
	return datum
}

func CloneBase(datum *types.Base) *types.Base {
	if datum == nil {
		return nil
	}
	clone := &types.Base{}
	clone.Active = datum.Active
	clone.Annotations = dataTest.CloneBlobArray(datum.Annotations)
	clone.Associations = dataTypesCommonAssociationTest.CloneAssociationArray(datum.Associations)
	clone.ArchivedDataSetID = test.CloneString(datum.ArchivedDataSetID)
	clone.ArchivedTime = test.CloneString(datum.ArchivedTime)
	clone.ClockDriftOffset = test.CloneInt(datum.ClockDriftOffset)
	clone.ConversionOffset = test.CloneInt(datum.ConversionOffset)
	clone.CreatedTime = test.CloneString(datum.CreatedTime)
	clone.CreatedUserID = test.CloneString(datum.CreatedUserID)
	clone.Deduplicator = dataTest.CloneDeduplicatorDescriptor(datum.Deduplicator)
	clone.DeletedTime = test.CloneString(datum.DeletedTime)
	clone.DeletedUserID = test.CloneString(datum.DeletedUserID)
	clone.DeviceID = test.CloneString(datum.DeviceID)
	clone.DeviceTime = test.CloneString(datum.DeviceTime)
	clone.GUID = test.CloneString(datum.GUID)
	clone.ID = test.CloneString(datum.ID)
	clone.Location = dataTypesCommonLocationTest.CloneLocation(datum.Location)
	clone.ModifiedTime = test.CloneString(datum.ModifiedTime)
	clone.ModifiedUserID = test.CloneString(datum.ModifiedUserID)
	clone.Notes = test.CloneStringArray(datum.Notes)
	clone.Origin = dataTypesCommonOriginTest.CloneOrigin(datum.Origin)
	clone.Payload = dataTest.CloneBlob(datum.Payload)
	clone.SchemaVersion = datum.SchemaVersion
	clone.Source = test.CloneString(datum.Source)
	clone.Tags = test.CloneStringArray(datum.Tags)
	clone.Time = test.CloneString(datum.Time)
	clone.TimeZoneName = test.CloneString(datum.TimeZoneName)
	clone.TimeZoneOffset = test.CloneInt(datum.TimeZoneOffset)
	clone.Type = datum.Type
	clone.UploadID = test.CloneString(datum.UploadID)
	clone.UserID = test.CloneString(datum.UserID)
	clone.Version = datum.Version
	return clone
}
