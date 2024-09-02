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
	//note: 'g6 pro' not specfied in API specs but found during actual usage
	DeviceTransmitterGenerationG6Pro  = "g6 pro"
	DeviceTransmitterGenerationG6Plus = "g6+"
	DeviceTransmitterGenerationPro    = "dexcomPro"
	DeviceTransmitterGenerationG7     = "g7"
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
		DeviceTransmitterGenerationG6Pro,
		DeviceTransmitterGenerationG6Plus,
		DeviceTransmitterGenerationPro,
		DeviceTransmitterGenerationG7,
	}
}

type DevicesResponse struct {
	RecordType    *string  `json:"recordType,omitempty"`
	RecordVersion *string  `json:"recordVersion,omitempty"`
	UserID        *string  `json:"userId,omitempty"`
	Devices       *Devices `json:"records,omitempty"`
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
	d.UserID = parser.String("userId")
	d.RecordType = parser.String("recordType")
	d.RecordVersion = parser.String("recordVersion")
	d.Devices = ParseDevices(parser.WithReferenceArrayParser("records"))
}

func (d *DevicesResponse) Validate(validator structure.Validator) {
	if devicesValidator := validator.WithReference("records"); d.Devices != nil {
		if !d.IsSandboxData {
			d.Devices.Validate(devicesValidator)
		}
	} else {
		devicesValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (d *DevicesResponse) Normalize(normalizer structure.Normalizer) {
	if d.Devices != nil {
		d.Devices.Normalize(normalizer.WithReference("records"))
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
	AlertScheduleList     *AlertSchedules `json:"alertSchedules,omitempty" yaml:"alertSchedules,omitempty"`
	TransmitterID         *string         `json:"transmitterId,omitempty" yaml:"transmitterId,omitempty"`
	TransmitterGeneration *string         `json:"transmitterGeneration,omitempty" yaml:"transmitterGeneration,omitempty"`
	DisplayDevice         *string         `json:"displayDevice,omitempty" yaml:"displayDevice,omitempty"`
	DisplayApp            *string         `json:"displayApp,omitempty" yaml:"displayApp,omitempty"`
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
	d.LastUploadDate = ParseTime(parser, "lastUploadDate")
	d.AlertScheduleList = ParseAlertSchedules(parser.WithReferenceArrayParser("alertSchedules"))
	d.TransmitterID = parser.String("transmitterId")
	d.TransmitterGeneration = parser.String("transmitterGeneration")
	d.DisplayDevice = parser.String("displayDevice")
	d.DisplayApp = parser.String("displayApp")
}

func (d *Device) Validate(validator structure.Validator) {
	validator = validator.WithMeta(d)
	validator.Time("lastUploadDate", d.LastUploadDate.Raw()).Exists().NotZero()
	if alertScheduleListValidator := validator.WithReference("alertSchedules"); d.AlertScheduleList != nil {
		d.AlertScheduleList.Validate(alertScheduleListValidator)
	} else {
		alertScheduleListValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.String("transmitterId", d.TransmitterID).Using(TransmitterIDValidator)
	validator.String("transmitterGeneration", d.TransmitterGeneration).Exists().OneOf(DeviceTransmitterGenerations()...)
	validator.String("displayDevice", d.DisplayDevice).Exists().OneOf(DeviceDisplayDevices()...)
}

func (d *Device) Normalize(normalizer structure.Normalizer) {
	if d.AlertScheduleList != nil {
		d.AlertScheduleList.Normalize(normalizer.WithReference("alertSchedules"))
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
