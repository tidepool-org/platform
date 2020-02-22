package pumpstatus

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "pumpStatus"

	MinReservoirRemaining = 0
	MaxReservoirRemaining = 1000
)

type PumpStatus struct {
	types.Base `bson:",inline"`

	Alerts             *[]string           `json:"alerts,omitempty" bson:"alerts,omitempty"`
	BasalDeliveryState *BasalDeliveryState `json:"basalDeliveryState,omitempty" bson:"basalDeliveryState,omitempty"`
	Battery            *Battery            `json:"battery,omitempty" bson:"battery,omitempty"`
	BolusState         *BolusState         `json:"bolusState,omitempty" bson:"bolusState,omitempty"`
	Forecast           *data.Forecast      `json:"forecast,omitempty" bson:"forecast,omitempty"`
	ReservoirRemaining *float64            `json:"reservoirRemaining,omitempty" bson:"reservoirRemaining,omitempty"`
	Device             *Device             `json:"device,omitempty" bson:"device,omitempty"`
}

func New() *PumpStatus {
	return &PumpStatus{
		Base: types.New(Type),
	}
}

func ParsePumpStatus(parser structure.ObjectParser) *PumpStatus {
	if !parser.Exists() {
		return nil
	}
	datum := NewPumpStatus()
	parser.Parse(datum)
	return datum
}

func NewPumpStatus() *PumpStatus {
	return &PumpStatus{}
}

func (c *PumpStatus) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(c.Meta())
	}

	c.Base.Parse(parser)

	c.Alerts = parser.StringArray("alerts")
	c.BasalDeliveryState = ParseBasalDeliveryState(parser.WithReferenceObjectParser("basalDeliveryState"))
	c.Battery = ParseBattery(parser.WithReferenceObjectParser("battery"))
	c.BolusState = ParseBolusState(parser.WithReferenceObjectParser("bolusState"))
	c.Device = ParseDevice(parser.WithReferenceObjectParser("device"))
	c.Forecast = data.ParseForecast(parser.WithReferenceObjectParser("forecast"))
	c.ReservoirRemaining = parser.Float64("reservoirRemaining")
}

func (c *PumpStatus) Validate(validator structure.Validator) {
	if c.Battery != nil {
		c.Battery.Validate(validator.WithReference("battery"))
	}
	if c.ReservoirRemaining != nil {
		validator.Float64("reservoirRemaining", c.ReservoirRemaining).InRange(MinReservoirRemaining, MaxReservoirRemaining)
	}
	if c.Forecast != nil {
		c.Forecast.Validate(validator.WithReference("forecast"))
	}
	if c.BasalDeliveryState != nil {
		c.BasalDeliveryState.Validate(validator.WithReference("basalDeliveryState"))
	}
	if c.BolusState != nil {
		c.BolusState.Validate(validator.WithReference("bolusState"))
	}
	if c.Device != nil {
		c.Device.Validate(validator.WithReference("device"))
	}
}

func (c *PumpStatus) Normalize(normalizer data.Normalizer) {
	if c.Battery != nil {
		c.Battery.Normalize(normalizer.WithReference("battery"))
	}
	if c.Forecast != nil {
		c.Forecast.Normalize(normalizer.WithReference("forecast"))
	}
	if c.BasalDeliveryState != nil {
		c.BasalDeliveryState.Normalize(normalizer.WithReference("basalDeliveryState"))
	}
	if c.BolusState != nil {
		c.BolusState.Normalize(normalizer.WithReference("bolusState"))
	}
	if c.Device != nil {
		c.Device.Normalize(normalizer.WithReference("device"))
	}
}
