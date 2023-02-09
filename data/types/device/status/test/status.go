package test

import (
	"math"

	"github.com/tidepool-org/platform/data/types/device/status"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewStatus() *status.Status {
	datum := status.New()
	datum.Device = *dataTypesDeviceTest.RandomDevice()
	datum.SubType = "status"
	datum.Duration = pointer.FromInt(test.RandomIntFromRange(status.DurationMinimum, math.MaxInt32))
	datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, math.MaxInt32))
	datum.Name = pointer.FromString(test.RandomStringFromArray(status.Names()))
	datum.Reason = metadataTest.RandomMetadata()
	return datum
}

func CloneStatus(datum *status.Status) *status.Status {
	if datum == nil {
		return nil
	}
	clone := status.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.Duration = pointer.CloneInt(datum.Duration)
	clone.DurationExpected = pointer.CloneInt(datum.DurationExpected)
	clone.Name = pointer.CloneString(datum.Name)
	clone.Reason = metadataTest.CloneMetadata(datum.Reason)
	return clone
}
