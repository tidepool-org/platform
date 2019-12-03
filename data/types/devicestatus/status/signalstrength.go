package status

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type SignalStrength struct {
	Unit  *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseSignalStrength(parser structure.ObjectParser) *SignalStrength {
	if !parser.Exists() {
		return nil
	}
	datum := NewSignalStrength()
	parser.Parse(datum)
	return datum
}

func NewSignalStrength() *SignalStrength {
	return &SignalStrength{}
}

func (s *SignalStrength) Parse(parser structure.ObjectParser) {
	s.Unit = parser.String("unit")
	s.Value = parser.Float64("value")
}

func (s *SignalStrength) Validate(validator structure.Validator) {
	validator.String("unit", s.Unit).Exists()
	validator.Float64("value", s.Value).Exists()
}

func (s *SignalStrength) Normalize(normalizer data.Normalizer) {
}
