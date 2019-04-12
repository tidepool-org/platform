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

func ParseFlight(parser structure.ObjectParser) *Flight {
	if !parser.Exists() {
		return nil
	}
	datum := NewFlight()
	parser.Parse(datum)
	return datum
}

func NewFlight() *Flight {
	return &Flight{}
}

func (f *Flight) Parse(parser structure.ObjectParser) {
	f.Count = parser.Int("count")
}

func (f *Flight) Validate(validator structure.Validator) {
	validator.Int("count", f.Count).Exists().InRange(FlightCountMinimum, FlightCountMaximum)
}

func (f *Flight) Normalize(normalizer data.Normalizer) {}
