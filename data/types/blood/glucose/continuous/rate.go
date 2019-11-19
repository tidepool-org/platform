package continuous

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	RateMaximum = 100
	RateMinimum = -100
)

type Rate struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func NewRate() *Rate {
	return &Rate{}
}

func (c *Rate) Parse(parser structure.ObjectParser) {

	c.Units = parser.String("units")
	c.Value = parser.Float64("value")
}

func (c *Rate) Validate(validator structure.Validator) {

	validator.Float64("value", c.Value).InRange(RateMinimum, RateMaximum)
}

func (c *Rate) Normalize(normalizer data.Normalizer) {
}

func ParseRate(parser structure.ObjectParser) *Rate {
	if !parser.Exists() {
		return nil
	}
	datum := NewRate()
	parser.Parse(datum)
	return datum
}
