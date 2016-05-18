package bloodglucose

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base"
)

type Continuous struct {
	base.Base `bson:",inline"`

	Value *float64 `json:"value" bson:"value"`
	Units *string  `json:"units" bson:"units"`
}

func ContinuousType() string {
	return "cbg"
}

func NewContinuous() *Continuous {
	bloodGlucoseType := ContinuousType()

	continuous := &Continuous{}
	continuous.Type = &bloodGlucoseType
	return continuous
}

func (c *Continuous) Parse(parser data.ObjectParser) {
	c.Base.Parse(parser)

	c.Value = parser.ParseFloat("value")
	c.Units = parser.ParseString("units")
}

func (c *Continuous) Validate(validator data.Validator) {
	c.Base.Validate(validator)

	validator.ValidateFloat("value", c.Value).Exists().InRange(0.0, 1000.0)
	validator.ValidateString("units", c.Units).Exists().OneOf([]string{"mmol/l", "mmol/L", "mg/dl", "mg/dL"})

}

func (c *Continuous) Normalize(normalizer data.Normalizer) {
	c.Base.Normalize(normalizer)
}
