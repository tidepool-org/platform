package oura

import (
	"github.com/tidepool-org/platform/structure"
)

type PersonalInfo struct {
	ID            *string  `json:"id,omitempty" bson:"id,omitempty"`
	Age           *int     `json:"age,omitempty" bson:"age,omitempty"`
	Weight        *float64 `json:"weight,omitempty" bson:"weight,omitempty"`
	Height        *float64 `json:"height,omitempty" bson:"height,omitempty"`
	BiologicalSex *string  `json:"biological_sex,omitempty" bson:"biological_sex,omitempty"`
	Email         *string  `json:"email,omitempty" bson:"email,omitempty"`
}

func (p *PersonalInfo) Parse(parser structure.ObjectParser) {
	p.ID = parser.String("id")
	p.Age = parser.Int("age")
	p.Weight = parser.Float64("weight")
	p.Height = parser.Float64("height")
	p.BiologicalSex = parser.String("biological_sex")
	p.Email = parser.String("email")
}

func (p *PersonalInfo) Validate(validator structure.Validator) {
	validator.String("id", p.ID).Exists().NotEmpty()
}
