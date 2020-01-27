package pumpstatus

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "pumpStatus"
)

type PumpStatus struct {
	types.Base `bson:",inline"`

	Alerts                     *[]string           `json:"alerts,omitempty" bson:"alerts,omitempty"`
	BasalDeliveryState         *string             `json:"basalDeliveryState,omitempty" bson:"basalDeliveryState,omitempty"`
	Battery                    *Battery            `json:"battery,omitempty" bson:"battery,omitempty"`
	BolusState                 *string             `json:"bolusState,omitempty" bson:"bolusState,omitempty"`
	Device                     *string             `json:"device,omitempty" bson:"device,omitempty"`
	Forecast                   *data.Forecast      `json:"forecast,omitempty" bson:"forecast,omitempty"`
	PumpBatteryChargeRemaining *float64            `json:"pumpBatteryChargeRemaing,omitempty" bson:"pumpBatteryChargeRemaing,omitempty"`
	ReservoirRemaining         *ReservoirRemaining `json:"reservoirRemaining,omitempty" bson:"reservoirRemaining,omitempty"`
	SignalStrength             *SignalStrength     `json:"signalStrength,omitempty" bson:"signalStrength,omitempty"`
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
	c.BasalDeliveryState = parser.String("basalDeliveryState")
	c.Battery = ParseBattery(parser.WithReferenceObjectParser("battery"))
	c.BolusState = parser.String("bolusState")
	c.Device = parser.String("device")
	c.Forecast = data.ParseForecast(parser.WithReferenceObjectParser("forecast"))
	c.PumpBatteryChargeRemaining = parser.Float64("pumpBatteryChargeRemaining")
	c.ReservoirRemaining = ParseReservoirRemaining(parser.WithReferenceObjectParser("reservoirRemaining"))
	c.SignalStrength = ParseSignalStrength(parser.WithReferenceObjectParser("signalStrength"))
}

func (c *PumpStatus) Validate(validator structure.Validator) {
	if c.Battery != nil {
		c.Battery.Validate(validator.WithReference("battery"))
	}
	if c.ReservoirRemaining != nil {
		c.ReservoirRemaining.Validate(validator.WithReference("reservoirRemaining"))
	}
	if c.SignalStrength != nil {
		c.SignalStrength.Validate(validator.WithReference("signalStrength"))
	}
	if c.Forecast != nil {
		c.Forecast.Validate(validator.WithReference("forecast"))
	}
}

func (c *PumpStatus) Normalize(normalizer data.Normalizer) {
	if c.Battery != nil {
		c.Battery.Normalize(normalizer.WithReference("battery"))
	}
	if c.SignalStrength != nil {
		c.SignalStrength.Normalize(normalizer.WithReference("signalStrength"))
	}
	if c.ReservoirRemaining != nil {
		c.ReservoirRemaining.Normalize(normalizer.WithReference("reservoirRemaining"))
	}
	if c.Forecast != nil {
		c.Forecast.Normalize(normalizer.WithReference("forecast"))
	}
}
