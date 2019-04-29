package test

import (
	dataTypesDeviceTimechange "github.com/tidepool-org/platform/data/types/device/timechange"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomChange() *dataTypesDeviceTimechange.Change {
	datum := dataTypesDeviceTimechange.NewChange()
	datum.Agent = pointer.FromString(test.RandomStringFromArray(dataTypesDeviceTimechange.Agents()))
	datum.From = pointer.FromString(test.RandomTime().Format(dataTypesDeviceTimechange.FromTimeFormat))
	datum.To = pointer.FromString(test.RandomTime().Format(dataTypesDeviceTimechange.ToTimeFormat))
	return datum
}

func CloneChange(datum *dataTypesDeviceTimechange.Change) *dataTypesDeviceTimechange.Change {
	if datum == nil {
		return nil
	}
	clone := dataTypesDeviceTimechange.NewChange()
	clone.Agent = test.CloneString(datum.Agent)
	clone.From = test.CloneString(datum.From)
	clone.To = test.CloneString(datum.To)
	return clone
}
