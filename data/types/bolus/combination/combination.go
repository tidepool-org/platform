package combination

import (
	"github.com/tidepool-org/platform/data"
	dataTypesBolus "github.com/tidepool-org/platform/data/types/bolus"
	dataTypesBolusExtended "github.com/tidepool-org/platform/data/types/bolus/extended"
	dataTypesBolusNormal "github.com/tidepool-org/platform/data/types/bolus/normal"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = "dual/square" // TODO: Rename Type to "bolus/combination"; remove SubType

	DurationMaximum = dataTypesBolusExtended.DurationMaximum
	DurationMinimum = dataTypesBolusExtended.DurationMinimum
	ExtendedMaximum = dataTypesBolusExtended.ExtendedMaximum
	ExtendedMinimum = dataTypesBolusExtended.ExtendedMinimum
	NormalMaximum   = dataTypesBolusNormal.NormalMaximum
	NormalMinimum   = dataTypesBolusNormal.NormalMinimum
)

type CombinationFields struct {
	dataTypesBolusExtended.ExtendedFields `bson:",inline"`
	dataTypesBolusNormal.NormalFields     `bson:",inline"`
}

func (c *CombinationFields) Parse(parser structure.ObjectParser) {
	c.ExtendedFields.Parse(parser)
	c.NormalFields.Parse(parser)
}

func (c *CombinationFields) Validate(validator structure.Validator) {
	c.ExtendedFields.Validate(validator)
	c.NormalFields.Validate(validator)
}

type Combination struct {
	dataTypesBolus.Bolus `bson:",inline"`

	CombinationFields `bson:",inline"`
}

func New() *Combination {
	return &Combination{
		Bolus: dataTypesBolus.New(SubType),
	}
}

func (c *Combination) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(c.Meta())
	}

	c.Bolus.Parse(parser)

	c.CombinationFields.Parse(parser)
}

func (c *Combination) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(c.Meta())
	}

	c.Bolus.Validate(validator)

	if c.SubType != "" {
		validator.String("subType", &c.SubType).EqualTo(SubType)
	}

	c.CombinationFields.Validate(validator)
}

func (c *Combination) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(c.Meta())
	}

	c.Bolus.Normalize(normalizer)
}
