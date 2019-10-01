package history

import (
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "history"
)

type History struct {
	Time    *string         `json:"time,omitempty" bson:"time,omitempty"`
	Changes *JSONPatchArray `json:"changes,omitempty" bson:"changes,omitempty"`
}

func New() *History {
	return &History{}
}

func ParseHistory(parser structure.ObjectParser) *History {
	if !parser.Exists() {
		return nil
	}
	datum := NewHistory()
	parser.Parse(datum)
	return datum
}

func NewHistory() *History {
	return &History{}
}

func (h *History) Parse(parser structure.ObjectParser) {
	h.Time = parser.String("time")
	h.Changes = ParseJSONPatchArray(parser.WithReferenceArrayParser("changes"))
}

func (h *History) Validate(validator structure.Validator, ref []byte) {
	if h.Changes != nil {
		h.Changes.Validate(validator, ref)
	}
}
