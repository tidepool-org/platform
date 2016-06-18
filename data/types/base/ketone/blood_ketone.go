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

func NewDatum() data.Datum {
	return New()
}

func New() *Blood {
	return &Blood{}
}

func Init() *Blood {
	blood := New()
	blood.Init()
	return blood
}

func (b *Blood) Init() {
	b.Base.Init()
	b.Base.Type = Type()

	b.Value = nil
	b.Units = nil
}

func (b *Blood) Parse(parser data.ObjectParser) error {
	parser.SetMeta(b.Meta())

	if err := b.Base.Parse(parser); err != nil {
		return err
	}

	b.Value = parser.ParseFloat("value")
	b.Units = parser.ParseString("units")

	return nil
}

func (b *Blood) Validate(validator data.Validator) error {
	validator.SetMeta(b.Meta())

	if err := b.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateStringAsBloodGlucoseUnits("units", b.Units).Exists()
	validator.ValidateFloatAsBloodGlucoseValue("value", b.Value).Exists().InRangeForUnits(b.Units)

	return nil
}

func (b *Blood) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(b.Meta())

	if err := b.Base.Normalize(normalizer); err != nil {
		return err
	}

	b.Units, b.Value = normalizer.NormalizeBloodGlucose(b.Units).UnitsAndValue(b.Value)

	return nil
}
