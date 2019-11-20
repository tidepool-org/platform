package transmitter

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = "deviceTransmitter" // TODO: Rename Type to "device/deviceTransmitter"; remove SubType

	Start   = "start"
	Stop    = "stop"
	Expired = "expired"
)

type Transmitter struct {
	device.Device `bson:",inline"`

	EventType *string `json:"status,omitempty" bson:"status,omitempty"`
}

func EventTypes() []string {
	return []string{Start, Stop, Expired}
}

func New() *Transmitter {
	return &Transmitter{
		Device: device.New(SubType),
	}
}

func (r *Transmitter) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(r.Meta())
	}

	r.Device.Parse(parser)
}

func (r *Transmitter) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(r.Meta())
	}

	r.Device.Validate(validator)

	if r.SubType != "" {
		validator.String("subType", &r.SubType).EqualTo(SubType)
	}

	validator.String("eventType", r.EventType).Exists().OneOf(EventTypes()...)
}

func (r *Transmitter) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(r.Meta())
	}

	r.Device.Normalize(normalizer)

}
