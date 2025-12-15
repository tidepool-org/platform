package dexcom

import (
	"crypto/md5"
	"encoding/hex"

	yaml "gopkg.in/yaml.v2"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	DevicesResponseRecordType    = "device"
	DevicesResponseRecordVersion = "3.0"

	DeviceDisplayDeviceUnknown             = "unknown"
	DeviceDisplayDeviceAndroid             = "android"
	DeviceDisplayDeviceIOS                 = "iOS"
	DeviceDisplayDeviceReceiver            = "receiver"
	DeviceDisplayDeviceShareReceiver       = "shareReceiver"
	DeviceDisplayDeviceTouchscreenReceiver = "touchscreenReceiver"

	DeviceDisplayAppUnknown  = "unknown"
	DeviceDisplayAppG5       = "G5"
	DeviceDisplayAppG6       = "G6"
	DeviceDisplayAppG7       = "G7"
	DeviceDisplayAppReceiver = "receiver"
	DeviceDisplayAppWatch    = "Watch"

	DeviceTransmitterGenerationUnknown = "unknown"
	DeviceTransmitterGenerationG4      = "g4"
	DeviceTransmitterGenerationG5      = "g5"
	DeviceTransmitterGenerationG6      = "g6"
	DeviceTransmitterGenerationG6Pro   = "g6 pro" // NOTE: Not specfied in API specs but found during actual usage
	DeviceTransmitterGenerationG6Plus  = "g6+"
	DeviceTransmitterGenerationPro     = "dexcomPro"
	DeviceTransmitterGenerationG7      = "g7"
	DeviceTransmitterGenerationG715Day = "g715day"
)

func DeviceDisplayDevices() []string {
	return []string{
		DeviceDisplayDeviceUnknown,
		DeviceDisplayDeviceAndroid,
		DeviceDisplayDeviceIOS,
		DeviceDisplayDeviceReceiver,
		DeviceDisplayDeviceShareReceiver,
		DeviceDisplayDeviceTouchscreenReceiver,
	}
}

func DeviceDisplayApps() []string {
	return []string{
		DeviceDisplayAppUnknown,
		DeviceDisplayAppG5,
		DeviceDisplayAppG6,
		DeviceDisplayAppG7,
		DeviceDisplayAppReceiver,
		DeviceDisplayAppWatch,
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
		DeviceTransmitterGenerationG715Day,
	}
}

type DevicesResponse struct {
	RecordType    *string  `json:"recordType,omitempty"`
	RecordVersion *string  `json:"recordVersion,omitempty"`
	UserID        *string  `json:"userId,omitempty"`
	Records       *Devices `json:"records,omitempty"`
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
	parser = parser.WithMeta(d)

	d.RecordType = parser.String("recordType")
	d.RecordVersion = parser.String("recordVersion")
	d.UserID = parser.String("userId")
	d.Records = ParseDevices(parser.WithReferenceArrayParser("records"))
}

func (d *DevicesResponse) Validate(validator structure.Validator) {
	validator = validator.WithMeta(d)

	validator.String("recordType", d.RecordType).Exists().EqualTo(DevicesResponseRecordType)
	validator.String("recordVersion", d.RecordVersion).Exists().EqualTo(DevicesResponseRecordVersion)
	validator.String("userId", d.UserID).Exists().NotEmpty()

	// Only validate that the records exist, remaining validation will occur later on a per-record basis
	if d.Records == nil {
		validator.WithReference("records").ReportError(structureValidator.ErrorValueNotExists())
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

type Device struct {
	LastUploadDate        *Time           `json:"lastUploadDate,omitempty" yaml:"-"`
	AlertSchedules        *AlertSchedules `json:"alertSchedules,omitempty" yaml:"alertSchedules,omitempty"`
	TransmitterGeneration *string         `json:"transmitterGeneration,omitempty" yaml:"transmitterGeneration,omitempty"`
	TransmitterID         *string         `json:"transmitterId,omitempty" yaml:"transmitterId,omitempty"`
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
	parser = parser.WithMeta(d)

	d.LastUploadDate = ParseTime(parser, "lastUploadDate")
	d.AlertSchedules = ParseAlertSchedules(parser.WithReferenceArrayParser("alertSchedules"))
	d.TransmitterGeneration = parser.String("transmitterGeneration")
	d.TransmitterID = parser.String("transmitterId")
	d.DisplayDevice = parser.String("displayDevice")
	d.DisplayApp = parser.String("displayApp")
}

func (d *Device) Validate(validator structure.Validator) {
	validator = validator.WithMeta(d)

	validator.Time("lastUploadDate", d.LastUploadDate.Raw()).NotZero() // Dexcom - May not exist
	if alertScheduleListValidator := validator.WithReference("alertSchedules"); d.AlertSchedules != nil {
		d.AlertSchedules.Validate(alertScheduleListValidator)
	} else {
		alertScheduleListValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.String("transmitterGeneration", d.TransmitterGeneration).Exists().OneOf(DeviceTransmitterGenerations()...)
	validator.String("transmitterId", d.TransmitterID).Exists().Using(TransmitterIDValidator)
	validator.String("displayDevice", d.DisplayDevice).Exists().OneOf(DeviceDisplayDevices()...)
	validator.String("displayApp", d.DisplayApp).Exists().OneOf(DeviceDisplayApps()...)
}

func (d *Device) Normalize(normalizer structure.Normalizer) {
	normalizer = normalizer.WithMeta(d)

	if d.AlertSchedules != nil {
		d.AlertSchedules.Normalize(normalizer.WithReference("alertSchedules"))
	}
}

func (d *Device) ID() string {
	if d.TransmitterID != nil && *d.TransmitterID != "" {
		return *d.TransmitterID
	} else {
		return *d.TransmitterGeneration
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
