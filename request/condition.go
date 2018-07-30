package request

import (
	"net/http"
	"strconv"

	"github.com/tidepool-org/platform/structure"
)

type Condition struct {
	Revision *int
}

func NewCondition() *Condition {
	return &Condition{}
}

func (c *Condition) Parse(parser structure.ObjectParser) {
	c.Revision = parser.Int("revision")
}

func (c *Condition) Validate(validator structure.Validator) {
	validator.Int("revision", c.Revision).GreaterThanOrEqualTo(0)
}

func (c *Condition) MutateRequest(req *http.Request) error {
	if c.Revision != nil {
		return NewParameterMutator("revision", strconv.Itoa(*c.Revision)).MutateRequest(req)
	}
	return nil
}
