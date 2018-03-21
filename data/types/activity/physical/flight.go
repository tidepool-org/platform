package physical

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	FlightCountMaximum = 10000
	FlightCountMinimum = 0
)

type Flight struct {
	Count *int `json:"count,omitempty" bson:"count,omitempty"`
}

func ParseFlight(parser data.ObjectParser) *Flight {
	if parser.Object() == nil {
		return nil
	}
	datum := NewFlight()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewFlight() *Flight {
	return &Flight{}
}

func (f *Flight) Parse(parser data.ObjectParser) {
	f.Count = parser.ParseInteger("count")
}

func (f *Flight) Validate(validator structure.Validator) {
	validator.Int("count", f.Count).Exists().InRange(FlightCountMinimum, FlightCountMaximum)
}

func (f *Flight) Normalize(normalizer data.Normalizer) {}
