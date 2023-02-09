package dosingdecision

import (
	"github.com/tidepool-org/platform/structure"
)

const (
	RecommendedBasalDurationMaximum = 86400000
	RecommendedBasalDurationMinimum = 0
	RecommendedBasalRateMaximum     = 100
	RecommendedBasalRateMinimum     = 0
)

type RecommendedBasal struct {
	Rate     *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
	Duration *int     `json:"duration,omitempty" bson:"duration,omitempty"`
}

func ParseRecommendedBasal(parser structure.ObjectParser) *RecommendedBasal {
	if !parser.Exists() {
		return nil
	}
	datum := NewRecommendedBasal()
	parser.Parse(datum)
	return datum
}

func NewRecommendedBasal() *RecommendedBasal {
	return &RecommendedBasal{}
}

func (r *RecommendedBasal) Parse(parser structure.ObjectParser) {
	r.Rate = parser.Float64("rate")
	r.Duration = parser.Int("duration")
}

func (r *RecommendedBasal) Validate(validator structure.Validator) {
	validator.Float64("rate", r.Rate).Exists().InRange(RecommendedBasalRateMinimum, RecommendedBasalRateMaximum)
	validator.Int("duration", r.Duration).InRange(RecommendedBasalDurationMinimum, RecommendedBasalDurationMaximum)
}
