package pump

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	InsulinDurationHoursMaximum = 10.0
	InsulinDurationHoursMinimum = 0.0
	InsulinUnitsHours           = "hours"
)

func InsulinUnits() []string {
	return []string{
		InsulinUnitsHours,
	}
}

type Insulin struct {
	Duration *float64 `json:"duration,omitempty" bson:"duration,omitempty"`
	Units    *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseInsulin(parser data.ObjectParser) *Insulin {
	if parser.Object() == nil {
		return nil
	}
	datum := NewInsulin()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewInsulin() *Insulin {
	return &Insulin{}
}

func (i *Insulin) Parse(parser data.ObjectParser) {
	i.Duration = parser.ParseFloat("duration")
	i.Units = parser.ParseString("units")
}

func (i *Insulin) Validate(validator structure.Validator) {
	validator.Float64("duration", i.Duration).Exists().InRange(InsulinDurationRangeForUnits(i.Units))
	validator.String("units", i.Units).Exists().OneOf(InsulinUnits()...)
}

func (i *Insulin) Normalize(normalizer data.Normalizer) {}

func InsulinDurationRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case InsulinUnitsHours:
			return InsulinDurationHoursMinimum, InsulinDurationHoursMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
