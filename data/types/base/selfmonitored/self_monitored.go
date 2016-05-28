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

func (b *BloodGlucose) Parse(parser data.ObjectParser) error {
	parser.SetMeta(b.Meta())

	if err := b.Base.Parse(parser); err != nil {
		return err
	}

	b.Value = parser.ParseFloat("value")
	b.Units = parser.ParseString("units")
	b.SubType = parser.ParseString("subType")

	return nil
}

func (b *BloodGlucose) Validate(validator data.Validator) error {
	validator.SetMeta(b.Meta())

	if err := b.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateStringAsBloodGlucoseUnits("units", b.Units).Exists()
	validator.ValidateFloatAsBloodGlucoseValue("value", b.Value).Exists().InRangeForUnits(b.Units)
	validator.ValidateString("subType", b.SubType).OneOf([]string{"manual", "linked"})

	return nil
}

func (b *BloodGlucose) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(b.Meta())

	if err := b.Base.Normalize(normalizer); err != nil {
		return err
	}

	b.Units, b.Value = normalizer.NormalizeBloodGlucose(b.Units).UnitsAndValue(b.Value)

	return nil
}
