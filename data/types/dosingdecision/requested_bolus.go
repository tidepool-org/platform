package dosingdecision

import (
	"github.com/tidepool-org/platform/structure"
)

const (
	RequestedBolusAmountMaximum = 1000
	RequestedBolusAmountMinimum = 0
)

type RequestedBolus struct {
	Amount *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
}

func ParseRequestedBolus(parser structure.ObjectParser) *RequestedBolus {
	if !parser.Exists() {
		return nil
	}
	datum := NewRequestedBolus()
	parser.Parse(datum)
	return datum
}

func NewRequestedBolus() *RequestedBolus {
	return &RequestedBolus{}
}

func (r *RequestedBolus) Parse(parser structure.ObjectParser) {
	r.Amount = parser.Float64("amount")
}

func (r *RequestedBolus) Validate(validator structure.Validator) {
	validator.Float64("amount", r.Amount).Exists().InRange(RequestedBolusAmountMinimum, RequestedBolusAmountMaximum)
}
