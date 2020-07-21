package pumpstatus

import (
	"time"

	"github.com/tidepool-org/platform/structure"
)

const (
	BatteryRemainingPercentMaximum = 100
	BatteryRemainingPercentMinimum = 0
	BatteryUnitsPercent            = "percent"
)

func BatteryUnits() []string {
	return []string{
		BatteryUnitsPercent,
	}
}

type Battery struct {
	Time      *time.Time `json:"time,omitempty" bson:"time,omitempty"`
	Remaining *float64   `json:"remaining,omitempty" bson:"remaining,omitempty"`
	Units     *string    `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseBattery(parser structure.ObjectParser) *Battery {
	if !parser.Exists() {
		return nil
	}
	datum := NewBattery()
	parser.Parse(datum)
	return datum
}

func NewBattery() *Battery {
	return &Battery{}
}

func (b *Battery) Parse(parser structure.ObjectParser) {
	b.Time = parser.Time("time", TimeFormat)
	b.Remaining = parser.Float64("remaining")
	b.Units = parser.String("units")
}

func (b *Battery) Validate(validator structure.Validator) {
	validator.Float64("remaining", b.Remaining).Exists().InRange(BatteryRemainingPercentMinimum, BatteryRemainingPercentMaximum)
	validator.String("units", b.Units).Exists().OneOf(BatteryUnits()...)
}
