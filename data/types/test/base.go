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

func RandomBase() *types.Base {
	createdTime := test.RandomTimeBefore(time.Now().Add(-30 * 24 * time.Hour))
	archivedTime := test.RandomTimeFromRange(createdTime, time.Now().Add(-7*24*time.Hour))
	modifiedTime := test.RandomTimeFromRange(archivedTime, time.Now().Add(-24*time.Hour))
	deletedTime := test.RandomTimeFromRange(modifiedTime, time.Now())

	datum := RandomBaseForParser()
	datum.ArchivedDataSetID = pointer.FromString(dataTest.RandomDataSetID())
	datum.ArchivedTime = pointer.FromTime(archivedTime)
	datum.ClockDriftOffset = pointer.FromInt(NewClockDriftOffset())
	datum.ConversionOffset = pointer.FromInt(NewConversionOffset())
	datum.CreatedTime = pointer.FromTime(createdTime)
	datum.CreatedUserID = pointer.FromString(userTest.RandomUserID())
	datum.Deduplicator = dataTest.RandomDeduplicatorDescriptor()
	datum.DeletedTime = pointer.FromTime(deletedTime)
	datum.DeletedUserID = pointer.FromString(userTest.RandomUserID())
	datum.DeviceID = pointer.FromString(dataTest.NewDeviceID())
	datum.DeviceTime = pointer.FromString(test.RandomTime().Format("2006-01-02T15:04:05"))
	datum.GUID = pointer.FromString(dataTest.RandomDatumID())
	datum.ID = pointer.FromString(dataTest.RandomDatumID())
	datum.Location = locationTest.RandomLocation()
	datum.ModifiedTime = pointer.FromTime(modifiedTime)
	datum.ModifiedUserID = pointer.FromString(userTest.RandomUserID())
	datum.Notes = pointer.FromStringArray([]string{NewNote(1, 20), NewNote(1, 20)})
	datum.Origin = originTest.RandomOrigin()
	datum.Payload = metadataTest.RandomMetadata()
	datum.Source = pointer.FromString("carelink")
	datum.Tags = pointer.FromStringArray([]string{NewTag(1, 10)})
	datum.Time = pointer.FromTime(test.RandomTime())
	datum.TimeZoneName = pointer.FromString(timeZoneTest.RandomName())
	datum.TimeZoneOffset = pointer.FromInt(NewTimeZoneOffset())
	datum.Type = NewType()
	datum.UploadID = pointer.FromString(dataTest.RandomDataSetID())
	datum.UserID = pointer.FromString(userTest.RandomUserID())
	datum.VersionInternal = NewVersionInternal()
	return datum
}

func RandomBaseForParser() *types.Base {
	datum := &types.Base{}
	datum.Active = false
	datum.Annotations = metadataTest.RandomMetadataArray()
	datum.Associations = associationTest.RandomAssociationArray()
	datum.ClockDriftOffset = pointer.FromInt(NewClockDriftOffset())
	datum.ConversionOffset = pointer.FromInt(NewConversionOffset())
	datum.DeviceID = pointer.FromString(dataTest.NewDeviceID())
	datum.DeviceTime = pointer.FromString(test.RandomTime().Format("2006-01-02T15:04:05"))
	datum.ID = pointer.FromString(dataTest.RandomDatumID())
	datum.Location = locationTest.RandomLocation()
	datum.Notes = pointer.FromStringArray([]string{NewNote(1, 20), NewNote(1, 20)})
	datum.Origin = originTest.RandomOrigin()
	datum.Payload = metadataTest.RandomMetadata()
	datum.Source = pointer.FromString("carelink")
	datum.Tags = pointer.FromStringArray([]string{NewTag(1, 10)})
	datum.Time = pointer.FromTime(test.RandomTime())
	datum.TimeZoneName = pointer.FromString(timeZoneTest.RandomName())
	datum.TimeZoneOffset = pointer.FromInt(NewTimeZoneOffset())
	datum.Type = NewType()
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
	clone.ArchivedTime = pointer.CloneTime(datum.ArchivedTime)
	clone.ClockDriftOffset = pointer.CloneInt(datum.ClockDriftOffset)
	clone.ConversionOffset = pointer.CloneInt(datum.ConversionOffset)
	clone.CreatedTime = pointer.CloneTime(datum.CreatedTime)
	clone.CreatedUserID = pointer.CloneString(datum.CreatedUserID)
	clone.Deduplicator = dataTest.CloneDeduplicatorDescriptor(datum.Deduplicator)
	clone.DeletedTime = pointer.CloneTime(datum.DeletedTime)
	clone.DeletedUserID = pointer.CloneString(datum.DeletedUserID)
	clone.DeviceID = pointer.CloneString(datum.DeviceID)
	clone.DeviceTime = pointer.CloneString(datum.DeviceTime)
	clone.GUID = pointer.CloneString(datum.GUID)
	clone.ID = pointer.CloneString(datum.ID)
	clone.Location = locationTest.CloneLocation(datum.Location)
	clone.ModifiedTime = pointer.CloneTime(datum.ModifiedTime)
	clone.ModifiedUserID = pointer.CloneString(datum.ModifiedUserID)
	clone.Notes = pointer.CloneStringArray(datum.Notes)
	clone.Origin = originTest.CloneOrigin(datum.Origin)
	clone.Payload = metadataTest.CloneMetadata(datum.Payload)
	clone.Source = pointer.CloneString(datum.Source)
	clone.Tags = pointer.CloneStringArray(datum.Tags)
	clone.Time = pointer.CloneTime(datum.Time)
	clone.TimeZoneName = pointer.CloneString(datum.TimeZoneName)
	clone.TimeZoneOffset = pointer.CloneInt(datum.TimeZoneOffset)
	clone.Type = datum.Type
	clone.UploadID = pointer.CloneString(datum.UploadID)
	clone.UserID = pointer.CloneString(datum.UserID)
	clone.VersionInternal = datum.VersionInternal
	return clone
}

func NewObjectFromBase(datum *types.Base, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if objectFormat == test.ObjectFormatBSON {
		object["_active"] = test.NewObjectFromBool(datum.Active, objectFormat)
	}
	if datum.Annotations != nil {
		object["annotations"] = metadataTest.NewArrayFromMetadataArray(datum.Annotations, objectFormat)
	}
	if datum.ArchivedDataSetID != nil {
		object["archivedDatasetId"] = test.NewObjectFromString(*datum.ArchivedDataSetID, objectFormat)
	}
	if datum.ArchivedTime != nil {
		object["archivedTime"] = test.NewObjectFromTime(*datum.ArchivedTime, objectFormat)
	}
	if datum.Associations != nil {
		object["associations"] = associationTest.NewArrayFromAssociationArray(datum.Associations, objectFormat)
	}
	if datum.ClockDriftOffset != nil {
		object["clockDriftOffset"] = test.NewObjectFromInt(*datum.ClockDriftOffset, objectFormat)
	}
	if datum.ConversionOffset != nil {
		object["conversionOffset"] = test.NewObjectFromInt(*datum.ConversionOffset, objectFormat)
	}
	if datum.CreatedTime != nil {
		object["createdTime"] = test.NewObjectFromTime(*datum.CreatedTime, objectFormat)
	}
	if datum.CreatedUserID != nil {
		object["createdUserId"] = test.NewObjectFromString(*datum.CreatedUserID, objectFormat)
	}
	if datum.Deduplicator != nil {
		if objectFormat == test.ObjectFormatBSON {
			object["_deduplicator"] = dataTest.NewObjectFromDeduplicatorDescriptor(datum.Deduplicator, objectFormat)
		} else {
			object["deduplicator"] = dataTest.NewObjectFromDeduplicatorDescriptor(datum.Deduplicator, objectFormat)
		}
	}
	if datum.DeletedTime != nil {
		object["deletedTime"] = test.NewObjectFromTime(*datum.DeletedTime, objectFormat)
	}
	if datum.DeletedUserID != nil {
		object["deletedUserId"] = test.NewObjectFromString(*datum.DeletedUserID, objectFormat)
	}
	if datum.DeviceID != nil {
		object["deviceId"] = test.NewObjectFromString(*datum.DeviceID, objectFormat)
	}
	if datum.DeviceTime != nil {
		object["deviceTime"] = test.NewObjectFromString(*datum.DeviceTime, objectFormat)
	}
	if datum.GUID != nil {
		object["guid"] = test.NewObjectFromString(*datum.GUID, objectFormat)
	}
	if datum.ID != nil {
		object["id"] = test.NewObjectFromString(*datum.ID, objectFormat)
	}
	if datum.Location != nil {
		object["location"] = locationTest.NewObjectFromLocation(datum.Location, objectFormat)
	}
	if datum.ModifiedTime != nil {
		object["modifiedTime"] = test.NewObjectFromTime(*datum.ModifiedTime, objectFormat)
	}
	if datum.ModifiedUserID != nil {
		object["modifiedUserId"] = test.NewObjectFromString(*datum.ModifiedUserID, objectFormat)
	}
	if datum.Notes != nil {
		object["notes"] = test.NewObjectFromStringArray(*datum.Notes, objectFormat)
	}
	if datum.Origin != nil {
		object["origin"] = originTest.NewObjectFromOrigin(datum.Origin, objectFormat)
	}
	if datum.Payload != nil {
		object["payload"] = metadataTest.NewObjectFromMetadata(datum.Payload, objectFormat)
	}
	if datum.Source != nil {
		object["source"] = test.NewObjectFromString(*datum.Source, objectFormat)
	}
	if datum.Tags != nil {
		object["tags"] = test.NewObjectFromStringArray(*datum.Tags, objectFormat)
	}
	if datum.Time != nil {
		object["time"] = test.NewObjectFromTime(*datum.Time, objectFormat)
	}
	if datum.TimeZoneName != nil {
		object["timezone"] = test.NewObjectFromString(*datum.TimeZoneName, objectFormat)
	}
	if datum.TimeZoneOffset != nil {
		object["timezoneOffset"] = test.NewObjectFromInt(*datum.TimeZoneOffset, objectFormat)
	}
	object["type"] = test.NewObjectFromString(datum.Type, objectFormat)
	if datum.UploadID != nil {
		object["uploadId"] = test.NewObjectFromString(*datum.UploadID, objectFormat)
	}
	if objectFormat == test.ObjectFormatBSON {
		if datum.UserID != nil {
			object["_userId"] = test.NewObjectFromString(*datum.UserID, objectFormat)
		}
		if datum.VersionInternal != 0 {
			object["_version"] = test.NewObjectFromInt(datum.VersionInternal, objectFormat)
		}
	}
	return object
}
