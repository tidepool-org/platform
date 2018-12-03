package timechange

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	AgentAutomatic = "automatic"
	AgentManual    = "manual"
	FromTimeFormat = "2006-01-02T15:04:05"
	ToTimeFormat   = "2006-01-02T15:04:05"
)

func Agents() []string {
	return []string{
		AgentAutomatic,
		AgentManual,
	}
}

type Change struct {
	Agent *string `json:"agent,omitempty" bson:"agent,omitempty"`
	From  *string `json:"from,omitempty" bson:"from,omitempty"`
	To    *string `json:"to,omitempty" bson:"to,omitempty"`
}

func ParseChange(parser data.ObjectParser) *Change {
	if parser.Object() == nil {
		return nil
	}
	datum := NewChange()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewChange() *Change {
	return &Change{}
}

func (c *Change) Parse(parser data.ObjectParser) {
	c.Agent = parser.ParseString("agent")
	c.From = parser.ParseString("from")
	c.To = parser.ParseString("to")
}

func (c *Change) Validate(validator structure.Validator) {
	validator.String("agent", c.Agent).Exists().OneOf(Agents()...)
	validator.String("from", c.From).Exists().AsTime(FromTimeFormat)
	validator.String("to", c.To).Exists().AsTime(ToTimeFormat)
}

func (c *Change) Normalize(normalizer data.Normalizer) {}
