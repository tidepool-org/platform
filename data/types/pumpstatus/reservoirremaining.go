package pumpstatus

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	MinReservoirAmount = 0
	MaxReservoirAmount = 1000
)

type ReservoirRemaining struct {
	Unit   *string  `json:"units,omitempty" bson:"units,omitempty"`
	Amount *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
}

func ParseReservoirRemaining(parser structure.ObjectParser) *ReservoirRemaining {
	if !parser.Exists() {
		return nil
	}
	datum := NewReservoirRemaining()
	parser.Parse(datum)
	return datum
}

func NewReservoirRemaining() *ReservoirRemaining {
	return &ReservoirRemaining{}
}

func (r *ReservoirRemaining) Parse(parser structure.ObjectParser) {
	r.Unit = parser.String("unit")
	r.Amount = parser.Float64("value")
}

func (r *ReservoirRemaining) Validate(validator structure.Validator) {
	validator.String("unit", r.Unit).Exists()
	validator.Float64("amount", r.Amount).Exists().InRange(MinReservoirAmount, MaxReservoirAmount)
}

func (r *ReservoirRemaining) Normalize(normalizer data.Normalizer) {
}
