package controller

import (
	"sort"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	FirmwareVersionLengthMaximum = 100
	HardwareVersionLengthMaximum = 100
	ManufacturerLengthMaximum    = 100
	ManufacturersLengthMaximum   = 10
	ModelLengthMaximum           = 100
	NameLengthMaximum            = 100
	SerialNumberLengthMaximum    = 100
	SoftwareVersionLengthMaximum = 100
)

type Device struct {
	FirmwareVersion *string   `json:"firmwareVersion,omitempty" bson:"firmwareVersion,omitempty"`
	HardwareVersion *string   `json:"hardwareVersion,omitempty" bson:"hardwareVersion,omitempty"`
	Manufacturers   *[]string `json:"manufacturers,omitempty" bson:"manufacturers,omitempty"`
	Model           *string   `json:"model,omitempty" bson:"model,omitempty"`
	Name            *string   `json:"name,omitempty" bson:"name,omitempty"`
	SerialNumber    *string   `json:"serialNumber,omitempty" bson:"serialNumber,omitempty"`
	SoftwareVersion *string   `json:"softwareVersion,omitempty" bson:"softwareVersion,omitempty"`
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
	d.FirmwareVersion = parser.String("firmwareVersion")
	d.HardwareVersion = parser.String("hardwareVersion")
	d.Manufacturers = parser.StringArray("manufacturers")
	d.Model = parser.String("model")
	d.Name = parser.String("name")
	d.SerialNumber = parser.String("serialNumber")
	d.SoftwareVersion = parser.String("softwareVersion")
}

func (d *Device) Validate(validator structure.Validator) {
	validator.String("firmwareVersion", d.FirmwareVersion).NotEmpty().LengthLessThanOrEqualTo(FirmwareVersionLengthMaximum)
	validator.String("hardwareVersion", d.HardwareVersion).NotEmpty().LengthLessThanOrEqualTo(HardwareVersionLengthMaximum)
	validator.StringArray("manufacturers", d.Manufacturers).NotEmpty().LengthLessThanOrEqualTo(ManufacturersLengthMaximum).Each(func(stringValidator structure.String) {
		stringValidator.Exists().NotEmpty().LengthLessThanOrEqualTo(ManufacturerLengthMaximum)
	}).EachUnique()
	validator.String("model", d.Model).NotEmpty().LengthLessThanOrEqualTo(ModelLengthMaximum)
	validator.String("name", d.Name).NotEmpty().LengthLessThanOrEqualTo(NameLengthMaximum)
	validator.String("serialNumber", d.SerialNumber).NotEmpty().LengthLessThanOrEqualTo(SerialNumberLengthMaximum)
	validator.String("softwareVersion", d.SoftwareVersion).NotEmpty().LengthLessThanOrEqualTo(SoftwareVersionLengthMaximum)

	if d.FirmwareVersion == nil && d.HardwareVersion == nil && d.Manufacturers == nil && d.Model == nil && d.Name == nil && d.SerialNumber == nil && d.SoftwareVersion == nil {
		validator.ReportError(structureValidator.ErrorValuesNotExistForAny("firmwareVersion", "hardwareVersion", "manufacturers", "model", "name", "serialNumber", "softwareVersion"))
	}
}

func (d *Device) Normalize(normalizer data.Normalizer) {
	if normalizer.Origin() == structure.OriginExternal {
		if d.Manufacturers != nil {
			sort.Strings(*d.Manufacturers)
		}
	}
}
