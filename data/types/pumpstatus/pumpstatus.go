package pumpstatus

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "pumpStatus"

	MinPumpChargeRemaining = 0.0
	MaxPumpChargeRemaining = 1.0
)

type PumpStatus struct {
	types.Base `bson:",inline"`

	Alerts             *[]string           `json:"alerts,omitempty" bson:"alerts,omitempty"`
	BasalDeliveryState *BasalDeliveryState `json:"basalDeliveryState,omitempty" bson:"basalDeliveryState,omitempty"`
	Battery            *Battery            `json:"battery,omitempty" bson:"battery,omitempty"`
	BolusState         *BolusState         `json:"bolusState,omitempty" bson:"bolusState,omitempty"`
	Device             *string             `json:"device,omitempty" bson:"device,omitempty"`
	Forecast           *data.Forecast      `json:"forecast,omitempty" bson:"forecast,omitempty"`
	ReservoirRemaining *ReservoirRemaining `json:"reservoirRemaining,omitempty" bson:"reservoirRemaining,omitempty"`
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
	c.Device = parser.String("device")
	c.Forecast = data.ParseForecast(parser.WithReferenceObjectParser("forecast"))
	c.ReservoirRemaining = ParseReservoirRemaining(parser.WithReferenceObjectParser("reservoirRemaining"))
}

func (c *PumpStatus) Validate(validator structure.Validator) {
	if c.Battery != nil {
		c.Battery.Validate(validator.WithReference("battery"))
	}
	if c.ReservoirRemaining != nil {
		c.ReservoirRemaining.Validate(validator.WithReference("reservoirRemaining"))
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
}

func (c *PumpStatus) Normalize(normalizer data.Normalizer) {
	if c.Battery != nil {
		c.Battery.Normalize(normalizer.WithReference("battery"))
	}
	if c.ReservoirRemaining != nil {
		c.ReservoirRemaining.Normalize(normalizer.WithReference("reservoirRemaining"))
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
}
