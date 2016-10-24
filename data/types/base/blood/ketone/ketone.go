package ketone

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base"
)

type Ketone struct {
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

func New() *Ketone {
	return &Ketone{}
}

func Init() *Ketone {
	ketone := New()
	ketone.Init()
	return ketone
}

func (k *Ketone) Init() {
	k.Base.Init()
	k.Base.Type = Type()

	k.Value = nil
	k.Units = nil
}

func (k *Ketone) Parse(parser data.ObjectParser) error {
	parser.SetMeta(k.Meta())

	if err := k.Base.Parse(parser); err != nil {
		return err
	}

	k.Value = parser.ParseFloat("value")
	k.Units = parser.ParseString("units")

	return nil
}

func (k *Ketone) Validate(validator data.Validator) error {
	validator.SetMeta(k.Meta())

	if err := k.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateStringAsBloodGlucoseUnits("units", k.Units).Exists()
	validator.ValidateFloatAsBloodGlucoseValue("value", k.Value).Exists().InRangeForUnits(k.Units)

	return nil
}

func (k *Ketone) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(k.Meta())

	if err := k.Base.Normalize(normalizer); err != nil {
		return err
	}

	k.Units, k.Value = normalizer.NormalizeBloodGlucose(k.Units).UnitsAndValue(k.Value)

	return nil
}
