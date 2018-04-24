package calculator

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	CarbohydrateMaximum = 100.0
	CarbohydrateMinimum = 0.0
	CorrectionMaximum   = 100.0
	CorrectionMinimum   = -100.0
	NetMaximum          = 100.0
	NetMinimum          = -100.0
)

type Recommended struct {
	Carbohydrate *float64 `json:"carb,omitempty" bson:"carb,omitempty"`
	Correction   *float64 `json:"correction,omitempty" bson:"correction,omitempty"`
	Net          *float64 `json:"net,omitempty" bson:"net,omitempty"`
}

func ParseRecommended(parser data.ObjectParser) *Recommended {
	if parser.Object() == nil {
		return nil
	}
	recommended := NewRecommended()
	recommended.Parse(parser)
	parser.ProcessNotParsed()
	return recommended
}

func NewRecommended() *Recommended {
	return &Recommended{}
}

func (r *Recommended) Parse(parser data.ObjectParser) {
	r.Carbohydrate = parser.ParseFloat("carb")
	r.Correction = parser.ParseFloat("correction")
	r.Net = parser.ParseFloat("net")
}

func (r *Recommended) Validate(validator structure.Validator) {
	validator.Float64("carb", r.Carbohydrate).InRange(CarbohydrateMinimum, CarbohydrateMaximum)
	validator.Float64("correction", r.Correction).InRange(CorrectionMinimum, CorrectionMaximum)
	validator.Float64("net", r.Net).InRange(NetMinimum, NetMaximum)
}

func (r *Recommended) Normalize(normalizer data.Normalizer) {}
