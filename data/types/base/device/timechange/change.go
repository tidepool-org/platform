package timechange

import "github.com/tidepool-org/platform/data"

type Change struct {
	From  *string `json:"from,omitempty" bson:"from,omitempty"`
	To    *string `json:"to,omitempty" bson:"to,omitempty"`
	Agent *string `json:"agent,omitempty" bson:"agent,omitempty"`
}

func NewChange() *Change {
	return &Change{}
}

func (c *Change) Parse(parser data.ObjectParser) {
	c.From = parser.ParseString("from")
	c.To = parser.ParseString("to")
	c.Agent = parser.ParseString("agent")
}

func (c *Change) Validate(validator data.Validator) {
	validator.ValidateStringAsTime("from", c.From, "2006-01-02T15:04:05").Exists()
	validator.ValidateStringAsTime("to", c.To, "2006-01-02T15:04:05").Exists()
	validator.ValidateString("agent", c.Agent).Exists().OneOf([]string{"manual", "automatic"})
}

func (c *Change) Normalize(normalizer data.Normalizer) {
}

func ParseChange(parser data.ObjectParser) *Change {
	var change *Change
	if parser.Object() != nil {
		change = NewChange()
		change.Parse(parser)
		parser.ProcessNotParsed()
	}
	return change
}
