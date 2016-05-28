package selfmonitored

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base"
)

type BloodGlucose struct {
	base.Base `bson:",inline"`

	Value   *float64 `json:"value,omitempty" bson:"value,omitempty"`
	Units   *string  `json:"units,omitempty" bson:"units,omitempty"`
	SubType *string  `json:"subType,omitempty" bson:"subType,omitempty"`
}

func Type() string {
	return "smbg"
}

func New() (*BloodGlucose, error) {
	bloodGlucoseBase, err := base.New(Type())
	if err != nil {
		return nil, err
	}

	return &BloodGlucose{
		Base: *bloodGlucoseBase,
	}, nil
}

func (b *BloodGlucose) Parse(parser data.ObjectParser) {
	parser.SetMeta(b.Meta())

	b.Base.Parse(parser)

	b.Value = parser.ParseFloat("value")
	b.Units = parser.ParseString("units")
	b.SubType = parser.ParseString("subType")
}

func (b *BloodGlucose) Validate(validator data.Validator) {
	validator.SetMeta(b.Meta())

	b.Base.Validate(validator)

	validator.ValidateStringAsBloodGlucoseUnits("units", b.Units).Exists()
	validator.ValidateFloatAsBloodGlucoseValue("value", b.Value).Exists().InRangeForUnits(b.Units)
	validator.ValidateString("subType", b.SubType).OneOf([]string{"manual", "linked"})
}

func (b *BloodGlucose) Normalize(normalizer data.Normalizer) {
	normalizer.SetMeta(b.Meta())

	b.Base.Normalize(normalizer)

	b.Units, b.Value = normalizer.NormalizeBloodGlucose(b.Units).UnitsAndValue(b.Value)
}
