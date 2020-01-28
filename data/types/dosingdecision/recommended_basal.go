package dosingdecision

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	MinUnitsPerHour = 0
	MaxUnitsPerHour = 1000
	MinDuration     = 0
	MaxDuration     = 1000
)

type RecommendedBasal struct {
	UnitsPerHour *float64 `json:"unitsPerHour,omitempty" bson:"unitsPerHour,omitempty"`
	Duration     *float64 `json:"duration,omitempty" bson:"duration,omitempty"`
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

func (i *RecommendedBasal) Parse(parser structure.ObjectParser) {
	i.UnitsPerHour = parser.Float64("unitsPerHour")
	i.Duration = parser.Float64("duration")
}

func (i *RecommendedBasal) Validate(validator structure.Validator) {
	validator.Float64("unitsPerHour", i.UnitsPerHour).Exists().InRange(MinUnitsPerHour, MaxUnitsPerHour)
	validator.Float64("duration", i.Duration).Exists().InRange(MinDuration, MaxDuration)
}

func (i *RecommendedBasal) Normalize(normalizer data.Normalizer) {
}
