package ketone

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base"
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
	parser.SetMeta(b.Meta())

	b.Base.Parse(parser)

	b.Value = parser.ParseFloat("value")
	b.Units = parser.ParseString("units")
}

func (b *Blood) Validate(validator data.Validator) {
	validator.SetMeta(b.Meta())

	b.Base.Validate(validator)

	validator.ValidateStringAsBloodGlucoseUnits("units", b.Units).Exists()
	validator.ValidateFloatAsBloodGlucoseValue("value", b.Value).Exists().InRangeForUnits(b.Units)
}

func (b *Blood) Normalize(normalizer data.Normalizer) {
	normalizer.SetMeta(b.Meta())

	b.Base.Normalize(normalizer)

	b.Units, b.Value = normalizer.NormalizeBloodGlucose(b.Units).UnitsAndValue(b.Value)
}
