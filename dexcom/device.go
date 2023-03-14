package dexcom

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"

	yaml "gopkg.in/yaml.v2"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	DeviceDisplayDeviceAndroid             = "android"
	DeviceDisplayDeviceIOS                 = "iOS"
	DeviceDisplayDeviceReceiver            = "receiver"
	DeviceDisplayDeviceShareReceiver       = "shareReceiver"
	DeviceDisplayDeviceTouchscreenReceiver = "touchscreenReceiver"

	DeviceTransmitterGenerationUnknown = "unknown"
	DeviceTransmitterGenerationG4      = "g4"
	DeviceTransmitterGenerationG5      = "g5"
	DeviceTransmitterGenerationG6      = "g6"
	DeviceTransmitterGenerationG6Pro   = "g6 pro"
	DeviceTransmitterGenerationG6Plus  = "g6+"
	DeviceTransmitterGenerationPro     = "dexcomPro"
	DeviceTransmitterGenerationG7      = "g7"
)

func DeviceDisplayDevices() []string {
	return []string{
		DeviceDisplayDeviceAndroid,
		DeviceDisplayDeviceIOS,
		DeviceDisplayDeviceReceiver,
		DeviceDisplayDeviceShareReceiver,
		DeviceDisplayDeviceTouchscreenReceiver,
	}
}

func DeviceTransmitterGenerations() []string {
	return []string{
		DeviceTransmitterGenerationUnknown,
		DeviceTransmitterGenerationG4,
		DeviceTransmitterGenerationG5,
		DeviceTransmitterGenerationG6,
		DeviceTransmitterGenerationG6Plus,
		DeviceTransmitterGenerationPro,
		DeviceTransmitterGenerationG7,
	}
}

type DevicesResponse struct {
	Devices       *Devices `json:"devices,omitempty"`
	IsSandboxData bool     `json:"isSandboxData,omitempty"`
}

func ParseDevicesResponse(parser structure.ObjectParser) *DevicesResponse {
	if !parser.Exists() {
		return nil
	}
	datum := NewDevicesResponse()
	parser.Parse(datum)
	return datum
}

func NewDevicesResponse() *DevicesResponse {
	return &DevicesResponse{}
}

func (d *DevicesResponse) Parse(parser structure.ObjectParser) {
	d.Devices = ParseDevices(parser.WithReferenceArrayParser("devices"))
}

func (d *DevicesResponse) Validate(validator structure.Validator) {
	if devicesValidator := validator.WithReference("devices"); d.Devices != nil {
		if !d.IsSandboxData {
			d.Devices.Validate(devicesValidator)
		}
	} else {
		devicesValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (d *DevicesResponse) Normalize(normalizer structure.Normalizer) {
	if d.Devices != nil {
		d.Devices.Normalize(normalizer.WithReference("devices"))
	}
}

type Devices []*Device

func ParseDevices(parser structure.ArrayParser) *Devices {
	if !parser.Exists() {
		return nil
	}
	datum := NewDevices()
	parser.Parse(datum)
	return datum
}

func NewDevices() *Devices {
	return &Devices{}
}

func (d *Devices) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*d = append(*d, ParseDevice(parser.WithReferenceObjectParser(reference)))
	}
}

func (d *Devices) Validate(validator structure.Validator) {
	for index, device := range *d {
		if deviceValidator := validator.WithReference(strconv.Itoa(index)); device != nil {
			device.Validate(deviceValidator)
		} else {
			deviceValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (d *Devices) Normalize(normalizer structure.Normalizer) {
	for index, device := range *d {
		device.Normalize(normalizer.WithReference(strconv.Itoa(index)))
	}
}

type Device struct {
	LastUploadDate        *Time           `json:"lastUploadDate,omitempty" yaml:"-"`
	AlertScheduleList     *AlertSchedules `json:"alertScheduleList,omitempty" yaml:"alertScheduleList,omitempty"`
	UDI                   *string         `json:"udi,omitempty" yaml:"udi,omitempty"`
	SerialNumber          *string         `json:"serialNumber,omitempty" yaml:"serialNumber,omitempty"`
	TransmitterID         *string         `json:"transmitterId,omitempty" yaml:"transmitterId,omitempty"`
	TransmitterGeneration *string         `json:"transmitterGeneration,omitempty" yaml:"transmitterGeneration,omitempty"`
	DisplayDevice         *string         `json:"displayDevice,omitempty" yaml:"displayDevice,omitempty"`
	SoftwareVersion       *string         `json:"softwareVersion,omitempty" yaml:"softwareVersion,omitempty"`
	SoftwareNumber        *string         `json:"softwareNumber,omitempty" yaml:"softwareNumber,omitempty"`
	Language              *string         `json:"language,omitempty" yaml:"language,omitempty"`
	IsMmolDisplayMode     *bool           `json:"isMmolDisplayMode,omitempty" yaml:"isMmolDisplayMode,omitempty"`
	IsBlindedMode         *bool           `json:"isBlindedMode,omitempty" yaml:"isBlindedMode,omitempty"`
	Is24HourMode          *bool           `json:"is24HourMode,omitempty" yaml:"is24HourMode,omitempty"`
	DisplayTimeOffset     *int            `json:"displayTimeOffset,omitempty" yaml:"displayTimeOffset,omitempty"`
	SystemTimeOffset      *int            `json:"systemTimeOffset,omitempty" yaml:"systemTimeOffset,omitempty"`
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
	d.LastUploadDate = TimeFromRaw(parser.Time("lastUploadDate", TimeFormat))
	d.AlertScheduleList = ParseAlertSchedules(parser.WithReferenceArrayParser("alertScheduleList"))
	d.UDI = parser.String("udi")
	d.SerialNumber = parser.String("serialNumber")
	d.TransmitterID = parser.String("transmitterId")
	d.TransmitterGeneration = parser.String("transmitterGeneration")
	d.DisplayDevice = parser.String("displayDevice")
	d.SoftwareVersion = parser.String("softwareVersion")
	d.SoftwareNumber = parser.String("softwareNumber")
	d.Language = parser.String("language")
	d.IsMmolDisplayMode = parser.Bool("isMmolDisplayMode")
	d.IsBlindedMode = parser.Bool("isBlindedMode")
	d.Is24HourMode = parser.Bool("is24HourMode")
	d.DisplayTimeOffset = parser.Int("displayTimeOffset")
	d.SystemTimeOffset = parser.Int("systemTimeOffset")
}

func (d *Device) Validate(validator structure.Validator) {
	validator = validator.WithMeta(d)
	validator.Time("lastUploadDate", d.LastUploadDate.Raw()).Exists().NotZero()
	if alertScheduleListValidator := validator.WithReference("alertScheduleList"); d.AlertScheduleList != nil {
		d.AlertScheduleList.Validate(alertScheduleListValidator)
	} else {
		alertScheduleListValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.String("udi", d.UDI).NotEmpty()
	validator.String("serialNumber", d.SerialNumber).Exists().NotEmpty()
	validator.String("transmitterId", d.TransmitterID).Using(TransmitterIDValidator)
	validator.String("transmitterGeneration", d.TransmitterGeneration).OneOf(DeviceTransmitterGenerations()...)
	validator.String("displayDevice", d.DisplayDevice).OneOf(DeviceDisplayDevices()...)
	validator.String("softwareVersion", d.SoftwareVersion).Exists().NotEmpty()
	validator.String("softwareNumber", d.SoftwareNumber).Exists().NotEmpty()
	validator.String("language", d.Language).Exists().NotEmpty()
}

func (d *Device) Normalize(normalizer structure.Normalizer) {
	if d.AlertScheduleList != nil {
		d.AlertScheduleList.Normalize(normalizer.WithReference("alertScheduleList"))
	}
}

func (d *Device) Hash() (string, error) {
	bites, err := yaml.Marshal(d)
	if err != nil {
		return "", errors.Wrap(err, "unable to generate hash")
	}
	md5Sum := md5.Sum(bites)
	return hex.EncodeToString(md5Sum[:]), nil
}
