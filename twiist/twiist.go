package twiist

import (
	"github.com/tidepool-org/platform/data"
)

const (
	DataSetClientName    = "com.sequelmedtech.tidepool-service"
	DataSetClientVersion = "3.0.0"
)

var (
	DeviceManufacturers = []string{"Sequel"}
	DeviceTags          = []string{data.DeviceTagBGM, data.DeviceTagCGM, data.DeviceTagInsulinPump}
)
