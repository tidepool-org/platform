package dosingdecision

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type Device struct {
	DeviceID     *string `json:"deviceID,omitempty" bson:"deviceID,omitempty"`
	Manufacturer *string `json:"manufacturer,omitempty" bson:"manufacturer,omitempty"`
	Model        *string `json:"model,omitempty" bson:"model,omitempty"`
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

func (i *Device) Parse(parser structure.ObjectParser) {
	i.DeviceID = parser.String("deviceID")
	i.Manufacturer = parser.String("manufacturer")
	i.Model = parser.String("model")
}

func (i *Device) Validate(validator structure.Validator) {
	validator.String("deviceID", i.DeviceID).Exists().LengthGreaterThanOrEqualTo(1)
	validator.String("model", i.Model).Exists().LengthGreaterThanOrEqualTo(1)
	validator.String("manufacturer", i.Manufacturer).Exists().LengthGreaterThanOrEqualTo(1)
}

func (i *Device) Normalize(normalizer data.Normalizer) {
}
