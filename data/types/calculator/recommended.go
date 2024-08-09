package calculator

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	CarbohydrateMaximum = 250.0
	CarbohydrateMinimum = 0.0
	CorrectionMaximum   = 250.0
	CorrectionMinimum   = -250.0
	NetMaximum          = 250.0
	NetMinimum          = -250.0
)

type Recommended struct {
	Carbohydrate *float64 `json:"carb,omitempty" bson:"carb,omitempty"`
	Correction   *float64 `json:"correction,omitempty" bson:"correction,omitempty"`
	Net          *float64 `json:"net,omitempty" bson:"net,omitempty"`
}

func ParseRecommended(parser structure.ObjectParser) *Recommended {
	if !parser.Exists() {
		return nil
	}
	datum := NewRecommended()
	parser.Parse(datum)
	return datum
}

func NewRecommended() *Recommended {
	return &Recommended{}
}

func (r *Recommended) Parse(parser structure.ObjectParser) {
	r.Carbohydrate = parser.Float64("carb")
	r.Correction = parser.Float64("correction")
	r.Net = parser.Float64("net")
}

func (r *Recommended) Validate(validator structure.Validator) {
	validator.Float64("carb", r.Carbohydrate).InRange(CarbohydrateMinimum, CarbohydrateMaximum)
	validator.Float64("correction", r.Correction).InRange(CorrectionMinimum, CorrectionMaximum)
	validator.Float64("net", r.Net).InRange(NetMinimum, NetMaximum)
}

func (r *Recommended) Normalize(normalizer data.Normalizer) {}
