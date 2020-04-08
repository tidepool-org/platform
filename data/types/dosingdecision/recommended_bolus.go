package dosingdecision

import (
	"github.com/tidepool-org/platform/structure"
)

const (
	RecommendedBolusAmountMaximum = 1000
	RecommendedBolusAmountMinimum = 0
)

type RecommendedBolus struct {
	Time   *string  `json:"time,omitempty" bson:"time,omitempty"`
	Amount *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
}

func ParseRecommendedBolus(parser structure.ObjectParser) *RecommendedBolus {
	if !parser.Exists() {
		return nil
	}
	datum := NewRecommendedBolus()
	parser.Parse(datum)
	return datum
}

func NewRecommendedBolus() *RecommendedBolus {
	return &RecommendedBolus{}
}

func (r *RecommendedBolus) Parse(parser structure.ObjectParser) {
	r.Time = parser.String("time")
	r.Amount = parser.Float64("amount")
}

func (r *RecommendedBolus) Validate(validator structure.Validator) {
	validator.String("time", r.Time).AsTime(TimeFormat)
	validator.Float64("amount", r.Amount).Exists().InRange(RecommendedBolusAmountMinimum, RecommendedBolusAmountMaximum)
}
