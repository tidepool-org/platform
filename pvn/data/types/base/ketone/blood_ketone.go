package ketone

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base"
	"github.com/tidepool-org/platform/pvn/data/types/common/bloodglucose"
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

	validator.ValidateString("units", b.Units).Exists().OneOf([]string{common.Mmoll, common.MmolL, common.Mgdl, common.MgdL})
	switch b.Units {
	case &common.Mmoll, &common.MmolL:
		validator.ValidateFloat("value", b.Value).Exists().InRange(common.MmolLFromValue, common.MmolLToValue)
	default:
		validator.ValidateFloat("value", b.Value).Exists().InRange(common.MgdLFromValue, common.MgdLToValue)
	}

}

func (b *Blood) Normalize(normalizer data.Normalizer) {
	b.Base.Normalize(normalizer)

	b.Units, b.Value = normalizer.NormalizeBloodGlucose(BloodType(), b.Units).NormalizeUnitsAndValue(b.Value)
}
