package device

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type       = "deviceEvent"
	StartEvent = "start"
	StopEvent  = "stop"
)

func Events() []string {
	return []string{
		StartEvent,
		StopEvent,
	}
}

type Device struct {
	types.Base `bson:",inline"`

	SubType   string  `json:"subType,omitempty" bson:"subType,omitempty"`
	EventType *string `json:"eventType,omitempty" bson:"eventType,omitempty"`
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

func NewWithEvent(subType string, eventType *string) Device {
	device := New(subType)
	device.EventType = pointer.CloneString(eventType)
	return device
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

	if d.EventType != nil {
		validator.String("eventType", d.EventType).OneOf(Events()...)
	}
}

func (d *Device) IdentityFields() ([]string, error) {
	identityFields, err := d.Base.IdentityFields()
	if err != nil {
		return nil, err
	}

	if d.SubType == "" {
		return nil, errors.New("sub type is empty")
	}

	return append(identityFields, d.SubType), nil
}
