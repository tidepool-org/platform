package ketone

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base"
)

type Blood struct {
	base.Base `bson:",inline"`

	Value *float64 `json:"value" bson:"value"`
	Units *string  `json:"units" bson:"units"`
}

func BloodType() string {
	return "bloodKetone"
}

func NewBlood() *Blood {
	ketoneType := BloodType()

	blood := &Blood{}
	blood.Type = &ketoneType
	return blood
}

func (b *Blood) Parse(parser data.ObjectParser) {
	b.Base.Parse(parser)

	b.Value = parser.ParseFloat("value")
	b.Units = parser.ParseString("units")
}

func (b *Blood) Validate(validator data.Validator) {
	b.Base.Validate(validator)

	validator.ValidateFloat("value", b.Value).Exists().InRange(0.0, 1000.0)
	validator.ValidateString("units", b.Units).Exists().OneOf([]string{"mmol/l", "mmol/L", "mg/dl", "mg/dL"})

}

func (b *Blood) Normalize(normalizer data.Normalizer) {
	b.Base.Normalize(normalizer)
}
