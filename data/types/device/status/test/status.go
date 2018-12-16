package test

import (
	"math"

	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device/status"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewStatus() *status.Status {
	datum := status.New()
	datum.Device = *dataTypesDeviceTest.NewDevice()
	datum.SubType = "status"
	datum.Duration = pointer.FromInt(test.RandomIntFromRange(status.DurationMinimum, math.MaxInt32))
	datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, math.MaxInt32))
	datum.Name = pointer.FromString(test.RandomStringFromArray(status.Names()))
	datum.Reason = dataTest.NewBlob()
	return datum
}

func CloneStatus(datum *status.Status) *status.Status {
	if datum == nil {
		return nil
	}
	clone := status.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.Duration = test.CloneInt(datum.Duration)
	clone.DurationExpected = test.CloneInt(datum.DurationExpected)
	clone.Name = test.CloneString(datum.Name)
	clone.Reason = dataTest.CloneBlob(datum.Reason)
	return clone
}
