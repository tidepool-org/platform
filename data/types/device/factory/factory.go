package factory

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	dataTypesDeviceAlarm "github.com/tidepool-org/platform/data/types/device/alarm"
	dataTypesDeviceCalibration "github.com/tidepool-org/platform/data/types/device/calibration"
	dataTypesDeviceParameter "github.com/tidepool-org/platform/data/types/device/deviceparameter"
	dataTypesDeviceFlush "github.com/tidepool-org/platform/data/types/device/flush"
	dataTypesDeviceMode "github.com/tidepool-org/platform/data/types/device/mode"
	dataTypesDevicePrime "github.com/tidepool-org/platform/data/types/device/prime"
	dataTypesDeviceReservoirchange "github.com/tidepool-org/platform/data/types/device/reservoirchange"
	dataTypesDeviceStatus "github.com/tidepool-org/platform/data/types/device/status"
	dataTypesDeviceTimechange "github.com/tidepool-org/platform/data/types/device/timechange"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var subTypes = []string{
	dataTypesDeviceAlarm.SubType,
	dataTypesDeviceCalibration.SubType,
	dataTypesDevicePrime.SubType,
	dataTypesDeviceReservoirchange.SubType,
	dataTypesDeviceStatus.SubType,
	dataTypesDeviceTimechange.SubType,
	dataTypesDeviceMode.ConfidentialMode,
	dataTypesDeviceMode.ZenMode,
	dataTypesDeviceMode.Warmup,
	dataTypesDeviceMode.LoopMode,
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
	case dataTypesDeviceFlush.SubType:
		return dataTypesDeviceFlush.New()
	case dataTypesDevicePrime.SubType:
		return dataTypesDevicePrime.New()
	case dataTypesDeviceReservoirchange.SubType:
		return dataTypesDeviceReservoirchange.New()
	case dataTypesDeviceStatus.SubType:
		return dataTypesDeviceStatus.New()
	case dataTypesDeviceTimechange.SubType:
		return dataTypesDeviceTimechange.New()
	case dataTypesDeviceParameter.SubType:
		return dataTypesDeviceParameter.New()
	case dataTypesDeviceMode.ConfidentialMode:
		return dataTypesDeviceMode.New(dataTypesDeviceMode.ConfidentialMode)
	case dataTypesDeviceMode.ZenMode:
		return dataTypesDeviceMode.New(dataTypesDeviceMode.ZenMode)
	case dataTypesDeviceMode.Warmup:
		return dataTypesDeviceMode.New(dataTypesDeviceMode.Warmup)
	case dataTypesDeviceMode.LoopMode:
		return dataTypesDeviceMode.New(dataTypesDeviceMode.LoopMode)
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
