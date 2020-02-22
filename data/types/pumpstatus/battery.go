package pumpstatus

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	MinBatteryPercentage = 0.0
	MaxBatteryPercentage = 100.0
)

type Battery struct {
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
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
	b.Value = parser.Float64("value")
}

func (b *Battery) Validate(validator structure.Validator) {
	validator.Float64("value", b.Value).Exists().InRange(MinBatteryPercentage, MaxBatteryPercentage)
}

func (b *Battery) Normalize(normalizer data.Normalizer) {
}
