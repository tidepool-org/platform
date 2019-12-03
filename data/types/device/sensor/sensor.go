package sensor

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = "deviceSensor" // TODO: Rename Type to "device/deviceSensor"; remove SubType

	Start   = "start"
	Stop    = "stop"
	Expired = "expired"
)

type Sensor struct {
	device.Device `bson:",inline"`

	EventType *string `json:"status,omitempty" bson:"status,omitempty"`
}

func EventTypes() []string {
	return []string{Start, Stop, Expired}
}

func New() *Sensor {
	return &Sensor{
		Device: device.New(SubType),
	}
}

func (r *Sensor) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(r.Meta())
	}

	r.Device.Parse(parser)
}

func (r *Sensor) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(r.Meta())
	}

	r.Device.Validate(validator)

	if r.SubType != "" {
		validator.String("subType", &r.SubType).EqualTo(SubType)
	}

	validator.String("eventType", r.EventType).Exists().OneOf(EventTypes()...)
}

func (r *Sensor) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(r.Meta())
	}

	r.Device.Normalize(normalizer)

}
