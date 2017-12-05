package ketone

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/blood/ketone"
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

	validator.ValidateString("units", k.Units).OneOf(ketone.Units())
	validator.ValidateFloat("value", k.Value).InRange(ketone.ValueRangeForUnits(k.Units))

	return nil
}

func (k *Ketone) Normalize(normalizer data.Normalizer) {
	normalizer = normalizer.WithMeta(k.Meta())

	k.Blood.Normalize(normalizer)

	k.Value = ketone.NormalizeValueForUnits(k.Value, k.Units)
	k.Units = ketone.NormalizeUnits(k.Units)
}
