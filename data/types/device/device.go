package device

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "deviceEvent"
)

type Device struct {
	types.Base `bson:",inline"`

	SubType string `json:"subType,omitempty" bson:"subType,omitempty"`
}

type Meta struct {
	Type    string `json:"type,omitempty"`
	SubType string `json:"subType,omitempty"`
}

func New(subType string) Device {
	return Device{
		Base:    types.New(Type),
		SubType: subType,
	}
}

func (d *Device) Meta() interface{} {
	return &Meta{
		Type:    d.Type,
		SubType: d.SubType,
	}
}

func (d *Device) Validate(validator structure.Validator) {
	d.Base.Validate(validator)

	if d.Type != "" {
		validator.String("type", &d.Type).EqualTo(Type)
	}

	validator.String("subType", &d.SubType).Exists().NotEmpty()
}

func (d *Device) IdentityFields(version string) ([]string, error) {
	identityFields, err := d.Base.IdentityFields(version)
	if err != nil {
		return nil, err
	}

	switch version {
	case types.IdentityFieldsVersionDeviceID, types.IdentityFieldsVersionDataSetID:
		if d.SubType == "" {
			return nil, errors.New("sub type is empty")
		}
		return append(identityFields, d.SubType), nil
	default:
		return nil, errors.New("version is invalid")
	}
}
