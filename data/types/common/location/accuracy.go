package location

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	AccuracyUnitsFeet          = "feet"
	AccuracyUnitsMeters        = "meters"
	AccuracyValueFeetMaximum   = AccuracyValueMetersMaximum / 0.3048
	AccuracyValueFeetMinimum   = AccuracyValueMetersMinimum / 0.3048
	AccuracyValueMetersMaximum = 1000.0
	AccuracyValueMetersMinimum = 0.0
)

func AccuracyUnits() []string {
	return []string{
		AccuracyUnitsFeet,
		AccuracyUnitsMeters,
	}
}

type Accuracy struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseAccuracy(parser structure.ObjectParser) *Accuracy {
	if !parser.Exists() {
		return nil
	}
	datum := NewAccuracy()
	parser.Parse(datum)
	return datum
}

func NewAccuracy() *Accuracy {
	return &Accuracy{}
}

func (a *Accuracy) Parse(parser structure.ObjectParser) {
	a.Units = parser.String("units")
	a.Value = parser.Float64("value")
}

func (a *Accuracy) Validate(validator structure.Validator) {
	validator.String("units", a.Units).Exists().OneOf(AccuracyUnits()...)
	validator.Float64("value", a.Value).Exists().InRange(AccuracyValueRangeForUnits(a.Units))
}

func (a *Accuracy) Normalize(normalizer data.Normalizer) {}

func AccuracyValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case AccuracyUnitsFeet:
			return AccuracyValueFeetMinimum, AccuracyValueFeetMaximum
		case AccuracyUnitsMeters:
			return AccuracyValueMetersMinimum, AccuracyValueMetersMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
