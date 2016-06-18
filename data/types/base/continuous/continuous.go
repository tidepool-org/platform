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

func NewDatum() data.Datum {
	return New()
}

func New() *BloodGlucose {
	return &BloodGlucose{}
}

func Init() *BloodGlucose {
	bloodGlucose := New()
	bloodGlucose.Init()
	return bloodGlucose
}

func (b *BloodGlucose) Init() {
	b.Base.Init()
	b.Base.Type = Type()

	b.Units = nil
	b.Value = nil
}

func (b *BloodGlucose) Parse(parser data.ObjectParser) error {
	parser.SetMeta(b.Meta())

	if err := b.Base.Parse(parser); err != nil {
		return err
	}

	b.Units = parser.ParseString("units")
	b.Value = parser.ParseFloat("value")

	return nil
}

func (b *BloodGlucose) Validate(validator data.Validator) error {
	validator.SetMeta(b.Meta())

	if err := b.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateStringAsBloodGlucoseUnits("units", b.Units).Exists()
	validator.ValidateFloatAsBloodGlucoseValue("value", b.Value).Exists().InRangeForUnits(b.Units)

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
