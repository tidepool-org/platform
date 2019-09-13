package status

import "github.com/tidepool-org/platform/structure"

type SignalStrength struct {
	Unit  *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

type SignalStrengthStruct struct {
	SignalStrength *SignalStrength `json:"signalStrength,omitempty" bson:"signalStrength,omitempty"`
}

func (s *SignalStrengthStruct) statusObject() {
}

func ParseSignalStrengthStruct(parser structure.ObjectParser) *SignalStrengthStruct {
	if !parser.Exists() {
		return nil
	}
	datum := NewSignalStrengthStruct()
	parser.Parse(datum)
	return datum
}

func NewSignalStrengthStruct() *SignalStrengthStruct {
	return &SignalStrengthStruct{}
}

func (s *SignalStrengthStruct) Parse(parser structure.ObjectParser) {
	s.SignalStrength = ParseSignalStrength(parser.WithReferenceObjectParser("battery"))
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
