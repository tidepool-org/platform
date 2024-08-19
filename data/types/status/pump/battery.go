package pump

import (
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	BatteryRemainingPercentMaximum = 1.0
	BatteryRemainingPercentMinimum = 0.0
	BatteryStateCharging           = "charging"
	BatteryStateFull               = "full"
	BatteryStateUnplugged          = "unplugged"
	BatteryUnitsPercent            = "percent"
)

func BatteryStates() []string {
	return []string{
		BatteryStateCharging,
		BatteryStateFull,
		BatteryStateUnplugged,
	}
}

func BatteryUnits() []string {
	return []string{
		BatteryUnitsPercent,
	}
}

type Battery struct {
	Time      *time.Time `json:"time,omitempty" bson:"time,omitempty"`
	State     *string    `json:"state,omitempty" bson:"state,omitempty"`
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
	b.Time = parser.Time("time", time.RFC3339Nano)
	b.State = parser.String("state")
	b.Remaining = parser.Float64("remaining")
	b.Units = parser.String("units")
}

func (b *Battery) Validate(validator structure.Validator) {
	validator.String("state", b.State).OneOf(BatteryStates()...)
	if b.Units != nil {
		switch *b.Units {
		case BatteryUnitsPercent:
			validator.Float64("remaining", b.Remaining).InRange(BatteryRemainingPercentMinimum, BatteryRemainingPercentMaximum)
		}
	}
	if unitsValidator := validator.String("units", b.Units); b.Remaining != nil {
		unitsValidator.Exists().OneOf(BatteryUnits()...)
	} else {
		unitsValidator.NotExists()
	}

	if b.State == nil && b.Remaining == nil {
		validator.ReportError(structureValidator.ErrorValuesNotExistForAny("state", "remaining"))
	}
}
