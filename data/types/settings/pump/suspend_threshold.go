package pump

import (
	"github.com/tidepool-org/platform/data"
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

func (b *SuspendThreshold) Parse(parser structure.ObjectParser) {
	b.Units = parser.String("units")
	b.Value = parser.Float64("value")
}

func (b *SuspendThreshold) Validate(validator structure.Validator) {
	validator.String("units", b.Units).Exists().OneOf(dataBloodGlucose.Units()...)
	validator.Float64("value", b.Value).Exists().InRange(dataBloodGlucose.ValueRangeForUnits(b.Units))
}

func (b *SuspendThreshold) Normalize(normalizer data.Normalizer) {}
