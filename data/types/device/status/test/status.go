package test

import (
	"math"

	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device/status"
	testDataTypesDevice "github.com/tidepool-org/platform/data/types/device/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewStatus() *status.Status {
	datum := status.New()
	datum.Device = *testDataTypesDevice.NewDevice()
	datum.SubType = "status"
	datum.Duration = pointer.Int(test.RandomIntFromRange(status.DurationMinimum, math.MaxInt32))
	datum.DurationExpected = pointer.Int(test.RandomIntFromRange(*datum.Duration, math.MaxInt32))
	datum.Name = pointer.String(test.RandomStringFromArray(status.Names()))
	datum.Reason = testData.NewBlob()
	return datum
}

func CloneStatus(datum *status.Status) *status.Status {
	if datum == nil {
		return nil
	}
	clone := status.New()
	clone.Device = *testDataTypesDevice.CloneDevice(&datum.Device)
	clone.Duration = test.CloneInt(datum.Duration)
	clone.DurationExpected = test.CloneInt(datum.DurationExpected)
	clone.Name = test.CloneString(datum.Name)
	clone.Reason = testData.CloneBlob(datum.Reason)
	return clone
}
