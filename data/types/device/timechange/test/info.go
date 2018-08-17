package test

import (
	dataTypesDeviceTimechange "github.com/tidepool-org/platform/data/types/device/timechange"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	timeZoneTest "github.com/tidepool-org/platform/time/zone/test"
)

func RandomInfo() *dataTypesDeviceTimechange.Info {
	datum := dataTypesDeviceTimechange.NewInfo()
	datum.Time = pointer.FromTime(test.RandomTime())
	datum.TimeZoneName = pointer.FromString(timeZoneTest.RandomName())
	return datum
}

func CloneInfo(datum *dataTypesDeviceTimechange.Info) *dataTypesDeviceTimechange.Info {
	if datum == nil {
		return nil
	}
	clone := dataTypesDeviceTimechange.NewInfo()
	clone.Time = test.CloneTime(datum.Time)
	clone.TimeZoneName = test.CloneString(datum.TimeZoneName)
	return clone
}
