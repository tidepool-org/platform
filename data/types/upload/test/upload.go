package test

import (
	dataTest "github.com/tidepool-org/platform/data/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

func RandomUpload() *dataTypesUpload.Upload {
	datum := dataTypesUpload.New()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "upload"
	datum.ByUser = pointer.FromString(userTest.RandomID())
	datum.Client = NewClient()
	datum.ComputerTime = pointer.FromString(test.RandomTime().Format("2006-01-02T15:04:05"))
	datum.DataSetType = pointer.FromString(test.RandomStringFromArray(dataTypesUpload.DataSetTypes()))
	datum.DataState = pointer.FromString(test.RandomStringFromArray(dataTypesUpload.States()))
	datum.Deduplicator = dataTest.RandomDeduplicatorDescriptor()
	datum.DeviceManufacturers = pointer.FromStringArray([]string{test.RandomStringFromRange(1, 16), test.RandomStringFromRange(1, 16)})
	datum.DeviceModel = pointer.FromString(test.RandomStringFromRange(1, 32))
	datum.DeviceSerialNumber = pointer.FromString(test.RandomStringFromRange(1, 16))
	datum.DeviceTags = pointer.FromStringArray(test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(dataTypesUpload.DeviceTags()), dataTypesUpload.DeviceTags()))
	datum.State = pointer.FromString(test.RandomStringFromArray(dataTypesUpload.States()))
	datum.TimeProcessing = pointer.FromString(dataTypesUpload.TimeProcessingUTCBootstrapping)
	datum.Version = pointer.FromString(netTest.RandomSemanticVersion())
	return datum
}

// this is an upload struct, with all time.Time fields replaced with string as they once were
type LegacyUpload struct {
	dataTypesTest.LegacyBase `bson:",inline"`

	ByUser              *string                 `json:"byUser,omitempty" bson:"byUser,omitempty"` // TODO: Deprecate in favor of CreatedUserID
	Client              *dataTypesUpload.Client `json:"client,omitempty" bson:"client,omitempty"`
	ComputerTime        *string                 `json:"computerTime,omitempty" bson:"computerTime,omitempty"` // TODO: Do we really need this? CreatedTime should suffice.
	DataSetType         *string                 `json:"dataSetType,omitempty" bson:"dataSetType,omitempty"`   // TODO: Migrate to "type" after migration to DataSet (not based on Base)
	DataState           *string                 `json:"-" bson:"_dataState,omitempty"`                        // TODO: Deprecated! (remove after data migration)
	DeviceManufacturers *[]string               `json:"deviceManufacturers,omitempty" bson:"deviceManufacturers,omitempty"`
	DeviceModel         *string                 `json:"deviceModel,omitempty" bson:"deviceModel,omitempty"`
	DeviceSerialNumber  *string                 `json:"deviceSerialNumber,omitempty" bson:"deviceSerialNumber,omitempty"`
	DeviceTags          *[]string               `json:"deviceTags,omitempty" bson:"deviceTags,omitempty"`
	State               *string                 `json:"-" bson:"_state,omitempty"` // TODO: Should this be returned in JSON? I think so.
	TimeProcessing      *string                 `json:"timeProcessing,omitempty" bson:"timeProcessing,omitempty"`
	Version             *string                 `json:"version,omitempty" bson:"version,omitempty"` // TODO: Deprecate in favor of Client.Version
}

func NewLegacy() *LegacyUpload {
	return &LegacyUpload{
		LegacyBase: dataTypesTest.NewLegacy("upload"),
	}
}

func RandomLegacyUpload() *LegacyUpload {
	datum := NewLegacy()
	datum.LegacyBase = *dataTypesTest.RandomLegacyBase()
	datum.Type = "upload"
	datum.ByUser = pointer.FromString(userTest.RandomID())
	datum.Client = NewClient()
	datum.ComputerTime = pointer.FromString(test.RandomTime().Format("2006-01-02T15:04:05"))
	datum.DataSetType = pointer.FromString(test.RandomStringFromArray(dataTypesUpload.DataSetTypes()))
	datum.DataState = pointer.FromString(test.RandomStringFromArray(dataTypesUpload.States()))
	datum.Deduplicator = dataTest.RandomDeduplicatorDescriptor()
	datum.DeviceManufacturers = pointer.FromStringArray([]string{test.RandomStringFromRange(1, 16), test.RandomStringFromRange(1, 16)})
	datum.DeviceModel = pointer.FromString(test.RandomStringFromRange(1, 32))
	datum.DeviceSerialNumber = pointer.FromString(test.RandomStringFromRange(1, 16))
	datum.DeviceTags = pointer.FromStringArray(test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(dataTypesUpload.DeviceTags()), dataTypesUpload.DeviceTags()))
	datum.State = pointer.FromString(test.RandomStringFromArray(dataTypesUpload.States()))
	datum.TimeProcessing = pointer.FromString(dataTypesUpload.TimeProcessingUTCBootstrapping)
	datum.Version = pointer.FromString(netTest.RandomSemanticVersion())
	return datum
}

func CloneUpload(datum *dataTypesUpload.Upload) *dataTypesUpload.Upload {
	if datum == nil {
		return nil
	}
	clone := dataTypesUpload.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.ByUser = pointer.CloneString(datum.ByUser)
	clone.Client = CloneClient(datum.Client)
	clone.ComputerTime = pointer.CloneString(datum.ComputerTime)
	clone.DataSetType = pointer.CloneString(datum.DataSetType)
	clone.DataState = pointer.CloneString(datum.DataState)
	clone.Deduplicator = dataTest.CloneDeduplicatorDescriptor(datum.Deduplicator)
	clone.DeviceManufacturers = pointer.CloneStringArray(datum.DeviceManufacturers)
	clone.DeviceModel = pointer.CloneString(datum.DeviceModel)
	clone.DeviceSerialNumber = pointer.CloneString(datum.DeviceSerialNumber)
	clone.DeviceTags = pointer.CloneStringArray(datum.DeviceTags)
	clone.State = pointer.CloneString(datum.State)
	clone.TimeProcessing = pointer.CloneString(datum.TimeProcessing)
	clone.Version = pointer.CloneString(datum.Version)
	return clone
}
