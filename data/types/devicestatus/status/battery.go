package status

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type Battery struct {
	Unit  *string  `json:"unit,omitempty" bson:"unit,omitempty"`
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
	b.Unit = parser.String("unit")
	b.Value = parser.Float64("value")
}

func (b *Battery) Validate(validator structure.Validator) {
	validator.String("unit", b.Unit).Exists()
	validator.Float64("value", b.Value).Exists()
}

func (b *Battery) Normalize(normalizer data.Normalizer) {
}
