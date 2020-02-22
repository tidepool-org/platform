package pumpstatus

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	MinDeviceIDLength     = 1
	MaxDeviceIDLength     = 1000
	MinModelLength        = 1
	MaxModelLength        = 1000
	MinManufacturerLength = 1
	MaxManufacturerLength = 1000
)

type Device struct {
	DeviceID     *string `json:"deviceID,omitempty" bson:"deviceID,omitempty"`
	Model        *string `json:"model,omitempty" bson:"model,omitempty"`
	Manufacturer *string `json:"manufacturer,omitempty" bson:"manufacturer,omitempty"`
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
func (b *Device) Parse(parser structure.ObjectParser) {
	b.DeviceID = parser.String("deviceID")
	b.Manufacturer = parser.String("manufacturer")
	b.Model = parser.String("model")
}

func (b *Device) Validate(validator structure.Validator) {
	validator.String("deviceID", b.DeviceID).Exists().LengthInRange(MinDeviceIDLength, MaxDeviceIDLength)
	validator.String("model", b.Model).Exists().LengthInRange(MinModelLength, MaxModelLength)
	validator.String("manufacturer", b.Manufacturer).Exists().LengthInRange(MinManufacturerLength, MaxManufacturerLength)
}

func (b *Device) Normalize(normalizer data.Normalizer) {
}
