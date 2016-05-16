package combination

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/bolus"
)

type Combination struct {
	bolus.Bolus

	Normal   *float64 `json:"normal" bson:"normal"`
	Duration *int     `json:"duration" bson:"duration"`
	Extended *float64 `json:"extended" bson:"extended"`
}

func Type() string {
	return bolus.Type()
}

func SubType() string {
	return "dual/square"
}

func New() *Combination {
	combinationType := Type()
	combinationSubType := SubType()

	combination := &Combination{}
	combination.Type = &combinationType
	combination.SubType = &combinationSubType
	return combination
}

func (c *Combination) Parse(parser data.ObjectParser) {
	c.Bolus.Parse(parser)
	c.Duration = parser.ParseInteger("duration")
	c.Extended = parser.ParseFloat("extended")
	c.Normal = parser.ParseFloat("normal")
}

func (c *Combination) Validate(validator data.Validator) {
	c.Bolus.Validate(validator)
	validator.ValidateInteger("duration", c.Duration).Exists().GreaterThanOrEqualTo(0).LessThanOrEqualTo(86400000)
	validator.ValidateFloat("extended", c.Extended).Exists().GreaterThan(0.0).LessThanOrEqualTo(100.0)
	validator.ValidateFloat("normal", c.Normal).Exists().GreaterThan(0.0).LessThanOrEqualTo(100.0)
}

func (c *Combination) Normalize(normalizer data.Normalizer) {
	c.Bolus.Normalize(normalizer)
}
