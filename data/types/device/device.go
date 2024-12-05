package device

import (
	"github.com/tidepool-org/platform/data/types"
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

func (d *Device) IdentityFields(version int) ([]string, error) {
	identityFields := []string{}
	var err error
	if version == types.LegacyIdentityFieldsVersion {

		identityFields, err = types.AppendIdentityFieldVal(identityFields, &d.Type, "type")
		if err != nil {
			return nil, err
		}

		identityFields, err = types.AppendIdentityFieldVal(identityFields, &d.SubType, "sub type")
		if err != nil {
			return nil, err
		}

		identityFields, err = types.AppendLegacyTimeVal(identityFields, d.Time)
		if err != nil {
			return nil, err
		}

		identityFields, err = types.AppendIdentityFieldVal(identityFields, d.DeviceID, "device id")
		if err != nil {
			return nil, err
		}

		return identityFields, nil
	}

	identityFields, err = d.Base.IdentityFields(version)
	if err != nil {
		return nil, err
	}
	identityFields, err = types.AppendIdentityFieldVal(identityFields, &d.SubType, "sub type")
	if err != nil {
		return nil, err
	}
	return identityFields, nil
}
