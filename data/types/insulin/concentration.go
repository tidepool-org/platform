package insulin

import (
	"math"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	ConcentrationUnitsUnitsPerML        = "Units/mL"
	ConcentrationValueUnitsPerMLMaximum = 10000.0
	ConcentrationValueUnitsPerMLMinimum = 0.0
)

func ConcentrationUnits() []string {
	return []string{
		ConcentrationUnitsUnitsPerML,
	}
}

type Concentration struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseConcentration(parser structure.ObjectParser) *Concentration {
	if !parser.Exists() {
		return nil
	}
	datum := NewConcentration()
	parser.Parse(datum)
	return datum
}

func NewConcentration() *Concentration {
	return &Concentration{}
}

func (c *Concentration) Parse(parser structure.ObjectParser) {
	c.Units = parser.String("units")
	c.Value = parser.Float64("value")
}

func (c *Concentration) Validate(validator structure.Validator) {
	validator.String("units", c.Units).Exists().OneOf(ConcentrationUnits()...)
	validator.Float64("value", c.Value).Exists().InRange(ConcentrationValueRangeForUnits(c.Units))
}

func (c *Concentration) Normalize(normalizer data.Normalizer) {}

func ConcentrationValueRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case ConcentrationUnitsUnitsPerML:
			return ConcentrationValueUnitsPerMLMinimum, ConcentrationValueUnitsPerMLMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
