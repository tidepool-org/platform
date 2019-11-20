package factory

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	dataTypesDeviceAlarm "github.com/tidepool-org/platform/data/types/device/alarm"
	dataTypesDeviceCalibration "github.com/tidepool-org/platform/data/types/device/calibration"
	dataTypesDevicePrime "github.com/tidepool-org/platform/data/types/device/prime"
	dataTypesDeviceReservoirchange "github.com/tidepool-org/platform/data/types/device/reservoirchange"
	dataTypesDeviceSensor "github.com/tidepool-org/platform/data/types/device/sensor"
	dataTypesDeviceStatus "github.com/tidepool-org/platform/data/types/device/status"
	dataTypesDeviceTimechange "github.com/tidepool-org/platform/data/types/device/timechange"
	dataTypesDeviceTransmitter "github.com/tidepool-org/platform/data/types/device/transmitter"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var subTypes = []string{
	dataTypesDeviceAlarm.SubType,
	dataTypesDeviceCalibration.SubType,
	dataTypesDevicePrime.SubType,
	dataTypesDeviceReservoirchange.SubType,
	dataTypesDeviceSensor.SubType,
	dataTypesDeviceStatus.SubType,
	dataTypesDeviceTimechange.SubType,
	dataTypesDeviceTransmitter.SubType,
}

func NewDeviceDatum(parser structure.ObjectParser) data.Datum {
	if !parser.Exists() {
		return nil
	}

	if value := parser.String("type"); value == nil {
		parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueNotExists())
		return nil
	} else if *value != device.Type {
		parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueStringNotOneOf(*value, []string{device.Type}))
		return nil
	}

	value := parser.String("subType")
	if value == nil {
		parser.WithReferenceErrorReporter("subType").ReportError(structureValidator.ErrorValueNotExists())
		return nil
	}

	switch *value {
	case dataTypesDeviceAlarm.SubType:
		return dataTypesDeviceAlarm.New()
	case dataTypesDeviceCalibration.SubType:
		return dataTypesDeviceCalibration.New()
	case dataTypesDevicePrime.SubType:
		return dataTypesDevicePrime.New()
	case dataTypesDeviceReservoirchange.SubType:
		return dataTypesDeviceReservoirchange.New()
	case dataTypesDeviceSensor.SubType:
		return dataTypesDeviceSensor.New()
	case dataTypesDeviceStatus.SubType:
		return dataTypesDeviceStatus.New()
	case dataTypesDeviceTimechange.SubType:
		return dataTypesDeviceTimechange.New()
	case dataTypesDeviceTransmitter.SubType:
		return dataTypesDeviceTransmitter.New()
	}

	parser.WithReferenceErrorReporter("subType").ReportError(structureValidator.ErrorValueStringNotOneOf(*value, subTypes))
	return nil
}

func ParseDeviceDatum(parser structure.ObjectParser) *data.Datum {
	datum := NewDeviceDatum(parser)
	if datum == nil {
		return nil
	}

	datum.Parse(parser)
	return &datum
}
