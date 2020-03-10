package pump

import (
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

type SuspendThreshold struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseSuspendThreshold(parser structure.ObjectParser) *SuspendThreshold {
	if !parser.Exists() {
		return nil
	}
	datum := NewSuspendThreshold()
	parser.Parse(datum)
	return datum
}

func NewSuspendThreshold() *SuspendThreshold {
	return &SuspendThreshold{}
}

func (s *SuspendThreshold) Parse(parser structure.ObjectParser) {
	s.Units = parser.String("units")
	s.Value = parser.Float64("value")
}

func (s *SuspendThreshold) Validate(validator structure.Validator) {
	validator.String("units", s.Units).Exists().OneOf(dataBloodGlucose.Units()...)
	validator.Float64("value", s.Value).Exists().InRange(dataBloodGlucose.ValueRangeForUnits(s.Units))
}
