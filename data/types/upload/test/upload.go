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
	datum.Base = *dataTypesTest.NewBase()
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
