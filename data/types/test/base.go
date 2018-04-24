package test

import (
	"time"

	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewBase() *types.Base {
	createdTime := test.NewTimeInRange(test.TimeMinimum(), time.Now().Add(-30*24*time.Hour))
	archivedTime := test.NewTimeInRange(createdTime, time.Now().Add(-7*24*time.Hour))
	modifiedTime := test.NewTimeInRange(archivedTime, time.Now().Add(-24*time.Hour))
	deletedTime := test.NewTimeInRange(modifiedTime, time.Now())

	datum := &types.Base{}
	datum.Active = false
	datum.Annotations = testData.NewBlobArray()
	datum.ArchivedDataSetID = pointer.String(id.New())
	datum.ArchivedTime = pointer.String(archivedTime.Format(time.RFC3339))
	datum.ClockDriftOffset = pointer.Int(NewClockDriftOffset())
	datum.ConversionOffset = pointer.Int(NewConversionOffset())
	datum.CreatedTime = pointer.String(createdTime.Format(time.RFC3339))
	datum.CreatedUserID = pointer.String(id.New())
	datum.Deduplicator = testData.NewDeduplicatorDescriptor()
	datum.DeletedTime = pointer.String(deletedTime.Format(time.RFC3339))
	datum.DeletedUserID = pointer.String(id.New())
	datum.DeviceID = pointer.String(id.New())
	datum.DeviceTime = pointer.String(test.NewTime().Format("2006-01-02T15:04:05"))
	datum.GUID = pointer.String(id.New())
	datum.ID = pointer.String(id.New())
	datum.ModifiedTime = pointer.String(modifiedTime.Format(time.RFC3339))
	datum.ModifiedUserID = pointer.String(id.New())
	datum.Payload = testData.NewBlob()
	datum.SchemaVersion = 2
	datum.Source = pointer.String("carelink")
	datum.Time = pointer.String(test.NewTime().Format(time.RFC3339))
	datum.TimezoneOffset = pointer.Int(NewTimezoneOffset())
	datum.Type = NewType()
	datum.UploadID = pointer.String(id.New())
	datum.UserID = pointer.String(id.New())
	datum.Version = NewVersion()
	return datum
}

func CloneBase(datum *types.Base) *types.Base {
	if datum == nil {
		return nil
	}
	clone := &types.Base{}
	clone.Active = datum.Active
	clone.Annotations = testData.CloneBlobArray(datum.Annotations)
	clone.ArchivedDataSetID = test.CloneString(datum.ArchivedDataSetID)
	clone.ArchivedTime = test.CloneString(datum.ArchivedTime)
	clone.ClockDriftOffset = test.CloneInt(datum.ClockDriftOffset)
	clone.ConversionOffset = test.CloneInt(datum.ConversionOffset)
	clone.CreatedTime = test.CloneString(datum.CreatedTime)
	clone.CreatedUserID = test.CloneString(datum.CreatedUserID)
	clone.Deduplicator = testData.CloneDeduplicatorDescriptor(datum.Deduplicator)
	clone.DeletedTime = test.CloneString(datum.DeletedTime)
	clone.DeletedUserID = test.CloneString(datum.DeletedUserID)
	clone.DeviceID = test.CloneString(datum.DeviceID)
	clone.DeviceTime = test.CloneString(datum.DeviceTime)
	clone.GUID = test.CloneString(datum.GUID)
	clone.ID = test.CloneString(datum.ID)
	clone.ModifiedTime = test.CloneString(datum.ModifiedTime)
	clone.ModifiedUserID = test.CloneString(datum.ModifiedUserID)
	clone.Payload = testData.CloneBlob(datum.Payload)
	clone.SchemaVersion = datum.SchemaVersion
	clone.Source = test.CloneString(datum.Source)
	clone.Time = test.CloneString(datum.Time)
	clone.TimezoneOffset = test.CloneInt(datum.TimezoneOffset)
	clone.Type = datum.Type
	clone.UploadID = test.CloneString(datum.UploadID)
	clone.UserID = test.CloneString(datum.UserID)
	clone.Version = datum.Version
	return clone
}
