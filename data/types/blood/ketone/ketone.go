package ketone

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/blood/ketone"
	"github.com/tidepool-org/platform/data/types/blood"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "bloodKetone"
)

type Ketone struct {
	blood.Blood `bson:",inline"`
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
	k.Type = Type
}

func (k *Ketone) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(k.Meta())
	}

	k.Blood.Validate(validator)

	if k.Type != "" {
		validator.String("type", &k.Type).EqualTo(Type)
	}

	validator.String("units", k.Units).Exists().OneOf(ketone.Units()...)
	validator.Float64("value", k.Value).Exists().InRange(ketone.ValueRangeForUnits(k.Units))
}

func (k *Ketone) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(k.Meta())
	}

	k.Blood.Normalize(normalizer)

	if normalizer.Origin() == structure.OriginExternal {
		units := k.Units
		k.Units = ketone.NormalizeUnits(units)
		k.Value = ketone.NormalizeValueForUnits(k.Value, units)
	}
}
