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

func New() *Ketone {
	return &Ketone{
		Blood: blood.New(Type),
	}
}

func (k *Ketone) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(k.Meta())
	}

	k.Blood.Parse(parser)
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

// IsValid returns true if there is no error in the validator
func (k *Ketone) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
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
