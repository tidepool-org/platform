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
	datum.ComputerTime = pointer.FromString(test.NewTime().Format("2006-01-02T15:04:05"))
	datum.DataSetType = pointer.FromString(test.RandomStringFromArray(dataTypesUpload.DataSetTypes()))
	datum.DataState = pointer.FromString(test.RandomStringFromArray(dataTypesUpload.States()))
	datum.Deduplicator = dataTest.RandomDeduplicatorDescriptor()
	datum.DeviceManufacturers = pointer.FromStringArray([]string{test.NewText(1, 16), test.NewText(1, 16)})
	datum.DeviceModel = pointer.FromString(test.NewText(1, 32))
	datum.DeviceSerialNumber = pointer.FromString(test.NewText(1, 16))
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
	clone.ByUser = test.CloneString(datum.ByUser)
	clone.Client = CloneClient(datum.Client)
	clone.ComputerTime = test.CloneString(datum.ComputerTime)
	clone.DataSetType = test.CloneString(datum.DataSetType)
	clone.DataState = test.CloneString(datum.DataState)
	clone.Deduplicator = dataTest.CloneDeduplicatorDescriptor(datum.Deduplicator)
	clone.DeviceManufacturers = test.CloneStringArray(datum.DeviceManufacturers)
	clone.DeviceModel = test.CloneString(datum.DeviceModel)
	clone.DeviceSerialNumber = test.CloneString(datum.DeviceSerialNumber)
	clone.DeviceTags = test.CloneStringArray(datum.DeviceTags)
	clone.State = test.CloneString(datum.State)
	clone.TimeProcessing = test.CloneString(datum.TimeProcessing)
	clone.Version = test.CloneString(datum.Version)
	return clone
}
