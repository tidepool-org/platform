package pump

import (
	"time"

	"github.com/tidepool-org/platform/data"
	dataTypes "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "pumpStatus"

	TimeFormat = time.RFC3339Nano
)

type Pump struct {
	dataTypes.Base `bson:",inline"`

	BasalDelivery *BasalDelivery `json:"basalDelivery,omitempty" bson:"basalDelivery,omitempty"`
	Battery       *Battery       `json:"battery,omitempty" bson:"battery,omitempty"`
	BolusDelivery *BolusDelivery `json:"bolusDelivery,omitempty" bson:"bolusDelivery,omitempty"`
	Device        *Device        `json:"device,omitempty" bson:"device,omitempty"`
	Reservoir     *Reservoir     `json:"reservoir,omitempty" bson:"reservoir,omitempty"`
}

func New() *Pump {
	return &Pump{
		Base: dataTypes.New(Type),
	}
}

func (p *Pump) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(p.Meta())
	}

	p.Base.Parse(parser)

	p.BasalDelivery = ParseBasalDelivery(parser.WithReferenceObjectParser("basalDelivery"))
	p.Battery = ParseBattery(parser.WithReferenceObjectParser("battery"))
	p.BolusDelivery = ParseBolusDelivery(parser.WithReferenceObjectParser("bolusDelivery"))
	p.Device = ParseDevice(parser.WithReferenceObjectParser("device"))
	p.Reservoir = ParseReservoir(parser.WithReferenceObjectParser("reservoir"))
}

func (p *Pump) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(p.Meta())
	}

	p.Base.Validate(validator)

	if p.Type != "" {
		validator.String("type", &p.Type).EqualTo(Type)
	}

	if p.BasalDelivery != nil {
		p.BasalDelivery.Validate(validator.WithReference("basalDelivery"))
	}
	if p.Battery != nil {
		p.Battery.Validate(validator.WithReference("battery"))
	}
	if p.BolusDelivery != nil {
		p.BolusDelivery.Validate(validator.WithReference("bolusDelivery"))
	}
	if p.Device != nil {
		p.Device.Validate(validator.WithReference("device"))
	}
	if p.Reservoir != nil {
		p.Reservoir.Validate(validator.WithReference("reservoir"))
	}
}

func (p *Pump) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(p.Meta())
	}

	p.Base.Normalize(normalizer)
}
