package ketone

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base"
	"github.com/tidepool-org/platform/pvn/data/types/common/bloodglucose"
)

type Blood struct {
	base.Base `bson:",inline"`

	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func Type() string {
	return "bloodKetone"
}

func New() (*Blood, error) {
	bloodBase, err := base.New(Type())
	if err != nil {
		return nil, err
	}

	return &Blood{
		Base: *bloodBase,
	}, nil
}

func (b *Blood) Parse(parser data.ObjectParser) {
	b.Base.Parse(parser)

	b.Value = parser.ParseFloat("value")
	b.Units = parser.ParseString("units")
}

func (b *Blood) Validate(validator data.Validator) {
	b.Base.Validate(validator)

	validator.ValidateString("units", b.Units).Exists().OneOf([]string{bloodglucose.Mmoll, bloodglucose.MmolL, bloodglucose.Mgdl, bloodglucose.MgdL})
	switch b.Units {
	case &bloodglucose.Mmoll, &bloodglucose.MmolL:
		validator.ValidateFloat("value", b.Value).Exists().InRange(bloodglucose.MmolLFromValue, bloodglucose.MmolLToValue)
	default:
		validator.ValidateFloat("value", b.Value).Exists().InRange(bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue)
	}
}

func (b *Blood) Normalize(normalizer data.Normalizer) {
	b.Base.Normalize(normalizer)

	b.Units, b.Value = normalizer.NormalizeBloodGlucose("value", b.Units).NormalizeUnitsAndValue(b.Value)
}
