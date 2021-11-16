package pump

import (
	"github.com/tidepool-org/platform/structure"
)

const (
	DeviceIDLengthMaximum           = 1000
	DeviceIDLengthMinimum           = 1
	DeviceNameLengthMaximum         = 1000
	DeviceNameLengthMinimum         = 1
	DeviceManufacturerLengthMaximum = 1000
	DeviceManufacturerLengthMinimum = 1
	DeviceModelLengthMaximum        = 1000
	DeviceModelLengthMinimum        = 1
	DeviceVersionLengthMaximum      = 100
	DeviceVersionLengthMinimum      = 1
)

type Device struct {
	ID              *string `json:"id,omitempty" bson:"id,omitempty"`
	Name            *string `json:"name,omitempty" bson:"name,omitempty"`
	Manufacturer    *string `json:"manufacturer,omitempty" bson:"manufacturer,omitempty"`
	Model           *string `json:"model,omitempty" bson:"model,omitempty"`
	FirmwareVersion *string `json:"firmwareVersion,omitempty" bson:"firmwareVersion,omitempty"`
	HardwareVersion *string `json:"hardwareVersion,omitempty" bson:"hardwareVersion,omitempty"`
	SoftwareVersion *string `json:"softwareVersion,omitempty" bson:"softwareVersion,omitempty"`
}

func ParseDevice(parser structure.ObjectParser) *Device {
	if !parser.Exists() {
		return nil
	}
	datum := NewDevice()
	parser.Parse(datum)
	return datum
}

func NewDevice() *Device {
	return &Device{}
}

func (d *Device) Parse(parser structure.ObjectParser) {
	d.ID = parser.String("id")
	d.Name = parser.String("name")
	d.Manufacturer = parser.String("manufacturer")
	d.Model = parser.String("model")
	d.FirmwareVersion = parser.String("firmwareVersion")
	d.HardwareVersion = parser.String("hardwareVersion")
	d.SoftwareVersion = parser.String("softwareVersion")
}

func (d *Device) Validate(validator structure.Validator) {
	validator.String("id", d.ID).LengthInRange(DeviceIDLengthMinimum, DeviceIDLengthMaximum)
	validator.String("name", d.Name).LengthInRange(DeviceNameLengthMinimum, DeviceNameLengthMaximum)
	validator.String("manufacturer", d.Manufacturer).LengthInRange(DeviceManufacturerLengthMinimum, DeviceManufacturerLengthMaximum)
	validator.String("model", d.Model).LengthInRange(DeviceModelLengthMinimum, DeviceModelLengthMaximum)
	validator.String("firmwareVersion", d.FirmwareVersion).LengthInRange(DeviceVersionLengthMinimum, DeviceVersionLengthMaximum)
	validator.String("hardwareVersion", d.HardwareVersion).LengthInRange(DeviceVersionLengthMinimum, DeviceVersionLengthMaximum)
	validator.String("softwareVersion", d.SoftwareVersion).LengthInRange(DeviceVersionLengthMinimum, DeviceVersionLengthMaximum)
}
