package pump

import (
	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

type BloodGlucosePreMealTarget struct {
	Low  *float64 `json:"low,omitempty" bson:"low,omitempty"`
	High *float64 `json:"high,omitempty" bson:"high,omitempty"`
}

func ParseBloodGlucosePreMealTarget(parser structure.ObjectParser) *BloodGlucosePreMealTarget {
	if !parser.Exists() {
		return nil
	}
	datum := NewBloodGlucosePreMealTarget()
	parser.Parse(datum)
	return datum
}

func NewBloodGlucosePreMealTarget() *BloodGlucosePreMealTarget {
	return &BloodGlucosePreMealTarget{}
}

func (b *BloodGlucosePreMealTarget) Parse(parser structure.ObjectParser) {
	b.Low = parser.Float64("low")
	b.High = parser.Float64("high")
}

func (b *BloodGlucosePreMealTarget) Validate(validator structure.Validator, units *string) {
	validator.Float64("low", b.Low).Exists().InRange(dataBloodGlucose.ValueRangeForUnits(units))
	validator.Float64("high", b.High).Exists().InRange(dataBloodGlucose.ValueRangeForUnits(units))
}

func (b *BloodGlucosePreMealTarget) Normalize(normalizer data.Normalizer, units *string) {
	if normalizer.Origin() == structure.OriginExternal {
		b.Low = dataBloodGlucose.NormalizeValueForUnits(b.Low, units)
		b.High = dataBloodGlucose.NormalizeValueForUnits(b.High, units)
	}
}
