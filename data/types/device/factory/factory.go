package factory

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	dataTypesDeviceAlarm "github.com/tidepool-org/platform/data/types/device/alarm"
	dataTypesDeviceCalibration "github.com/tidepool-org/platform/data/types/device/calibration"
	dataTypesDevicePrime "github.com/tidepool-org/platform/data/types/device/prime"
	dataTypesDeviceReservoirchange "github.com/tidepool-org/platform/data/types/device/reservoirchange"
	dataTypesDeviceStatus "github.com/tidepool-org/platform/data/types/device/status"
	dataTypesDeviceTimechange "github.com/tidepool-org/platform/data/types/device/timechange"
	"github.com/tidepool-org/platform/service"
)

var subTypes = []string{
	dataTypesDeviceAlarm.SubType,
	dataTypesDeviceCalibration.SubType,
	dataTypesDevicePrime.SubType,
	dataTypesDeviceReservoirchange.SubType,
	dataTypesDeviceStatus.SubType,
	dataTypesDeviceTimechange.SubType,
}

func NewDeviceDatum(parser data.ObjectParser) data.Datum {
	if parser.Object() == nil {
		return nil
	}

	if value := parser.ParseString("type"); value == nil {
		parser.AppendError("type", service.ErrorValueNotExists())
		return nil
	} else if *value != device.Type {
		parser.AppendError("type", service.ErrorValueStringNotOneOf(*value, []string{device.Type}))
		return nil
	}

	value := parser.ParseString("subType")
	if value == nil {
		parser.AppendError("subType", service.ErrorValueNotExists())
		return nil
	}

	switch *value {
	case dataTypesDeviceAlarm.SubType:
		return dataTypesDeviceAlarm.Init()
	case dataTypesDeviceCalibration.SubType:
		return dataTypesDeviceCalibration.Init()
	case dataTypesDevicePrime.SubType:
		return dataTypesDevicePrime.Init()
	case dataTypesDeviceReservoirchange.SubType:
		return dataTypesDeviceReservoirchange.Init()
	case dataTypesDeviceStatus.SubType:
		return dataTypesDeviceStatus.Init()
	case dataTypesDeviceTimechange.SubType:
		return dataTypesDeviceTimechange.Init()
	}

	parser.AppendError("subType", service.ErrorValueStringNotOneOf(*value, subTypes))
	return nil
}

func ParseDeviceDatum(parser data.ObjectParser) *data.Datum {
	datum := NewDeviceDatum(parser)
	if datum == nil {
		return nil
	}

	datum.Parse(parser)
	return &datum
}
