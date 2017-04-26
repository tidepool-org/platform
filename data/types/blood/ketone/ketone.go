package ketone

import (
	"github.com/tidepool-org/platform/data"
	commonKetone "github.com/tidepool-org/platform/data/blood/ketone"
	"github.com/tidepool-org/platform/data/types/blood"
)

type Ketone struct {
	blood.Blood `bson:",inline"`
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
	k.Blood.Init()
	k.Type = Type()
}

func (k *Ketone) Validate(validator data.Validator) error {
	if err := k.Blood.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("type", &k.Type).EqualTo(Type())

	validator.ValidateString("units", k.Units).OneOf(commonKetone.Units())
	validator.ValidateFloat("value", k.Value).InRange(commonKetone.ValueRangeForUnits(k.Units))

	return nil
}

func (k *Ketone) Normalize(normalizer data.Normalizer) error {
	if err := k.Blood.Normalize(normalizer); err != nil {
		return err
	}

	k.Value = commonKetone.NormalizeValueForUnits(k.Value, k.Units)
	k.Units = commonKetone.NormalizeUnits(k.Units)

	return nil
}
