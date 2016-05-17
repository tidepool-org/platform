package calculator

import "github.com/tidepool-org/platform/pvn/data"

type Recommended struct {
	Carbohydrate *float64 `json:"carb" bson:"carb"`
	Correction   *float64 `json:"correction" bson:"correction"`
	Net          *float64 `json:"net" bson:"net"`
}

func NewRecommended() *Recommended {
	return &Recommended{}
}

func (r *Recommended) Parse(parser data.ObjectParser) {
	r.Carbohydrate = parser.ParseFloat("carb")
	r.Correction = parser.ParseFloat("correction")
	r.Net = parser.ParseFloat("net")
}

func (r *Recommended) Validate(validator data.Validator) {
	validator.ValidateFloat("carb", r.Carbohydrate).GreaterThanOrEqualTo(0.0).LessThanOrEqualTo(100.0)
	validator.ValidateFloat("correction", r.Correction).GreaterThanOrEqualTo(-100.0).LessThanOrEqualTo(100.0)
	validator.ValidateFloat("net", r.Net).GreaterThanOrEqualTo(-100.0).LessThanOrEqualTo(100.0)
}

func (r *Recommended) Normalize(normalizer data.Normalizer) {
}

func ParseRecommended(parser data.ObjectParser) *Recommended {
	var recommended *Recommended
	if parser.Object() != nil {
		recommended = NewRecommended()
		recommended.Parse(parser)
	}
	return recommended
}
