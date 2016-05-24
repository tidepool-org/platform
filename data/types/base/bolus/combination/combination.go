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
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base/bolus"
)

type Combination struct {
	bolus.Bolus `bson:",inline"`

	Normal   *float64 `json:"normal,omitempty" bson:"normal,omitempty"`
	Duration *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	Extended *float64 `json:"extended,omitempty" bson:"extended,omitempty"`
}

func SubType() string {
	return "dual/square"
}

func New() (*Combination, error) {
	combinationBolus, err := bolus.New(SubType())
	if err != nil {
		return nil, err
	}

	return &Combination{
		Bolus: *combinationBolus,
	}, nil
}

func (c *Combination) Parse(parser data.ObjectParser) {
	c.Bolus.Parse(parser)

	c.Duration = parser.ParseInteger("duration")
	c.Extended = parser.ParseFloat("extended")
	c.Normal = parser.ParseFloat("normal")
}

func (c *Combination) Validate(validator data.Validator) {
	c.Bolus.Validate(validator)

	validator.ValidateInteger("duration", c.Duration).Exists().InRange(0, 86400000)
	validator.ValidateFloat("extended", c.Extended).Exists().InRange(0.0, 100.0)
	validator.ValidateFloat("normal", c.Normal).Exists().GreaterThan(0.0).LessThanOrEqualTo(100.0)
}
