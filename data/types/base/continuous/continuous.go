package continuous

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base"
)

type BloodGlucose struct {
	base.Base `bson:",inline"`

	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func Type() string {
	return "cbg"
}

func New() (*BloodGlucose, error) {
	continuousBase, err := base.New(Type())
	if err != nil {
		return nil, err
	}

	return &BloodGlucose{
		Base: *continuousBase,
	}, nil
}

func (b *BloodGlucose) Parse(parser data.ObjectParser) {
	parser.SetMeta(b.Meta())

	b.Base.Parse(parser)

	b.Units = parser.ParseString("units")
	b.Value = parser.ParseFloat("value")
}

func (b *BloodGlucose) Validate(validator data.Validator) {
	validator.SetMeta(b.Meta())

	b.Base.Validate(validator)

	validator.ValidateStringAsBloodGlucoseUnits("units", b.Units).Exists()
	validator.ValidateFloatAsBloodGlucoseValue("value", b.Value).Exists().InRangeForUnits(b.Units)
}

func (b *BloodGlucose) Normalize(normalizer data.Normalizer) {
	normalizer.SetMeta(b.Meta())

	b.Base.Normalize(normalizer)

	b.Units, b.Value = normalizer.NormalizeBloodGlucose(b.Units).UnitsAndValue(b.Value)
}
