package structured

import (
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
)

type Condition struct {
	Revision *int `json:"revision,omitempty"`
}

func NewCondition() *Condition {
	return &Condition{}
}

func MapCondition(condition *request.Condition) *Condition {
	if condition == nil {
		return nil
	}
	return &Condition{
		Revision: condition.Revision,
	}
}

func (c *Condition) Parse(parser structure.ObjectParser) {
	c.Revision = parser.Int("revision")
}

func (c *Condition) Validate(validator structure.Validator) {
	validator.Int("revision", c.Revision).GreaterThanOrEqualTo(0)
}
