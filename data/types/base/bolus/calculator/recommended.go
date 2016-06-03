package calculator

import "github.com/tidepool-org/platform/data"

type Recommended struct {
	Carbohydrate *float64 `json:"carb,omitempty" bson:"carb,omitempty"`
	Correction   *float64 `json:"correction,omitempty" bson:"correction,omitempty"`
	Net          *float64 `json:"net,omitempty" bson:"net,omitempty"`
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
	validator.ValidateFloat("carb", r.Carbohydrate).InRange(0.0, 100.0)
	validator.ValidateFloat("correction", r.Correction).InRange(-100.0, 100.0)
	validator.ValidateFloat("net", r.Net).InRange(-100.0, 100.0)
}

func (r *Recommended) Normalize(normalizer data.Normalizer) {
}

func ParseRecommended(parser data.ObjectParser) *Recommended {
	var recommended *Recommended
	if parser.Object() != nil {
		recommended = NewRecommended()
		recommended.Parse(parser)
		parser.ProcessNotParsed()
	}
	return recommended
}
