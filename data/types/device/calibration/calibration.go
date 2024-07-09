package calibration

import (
	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = "calibration" // TODO: Rename Type to "device/calibration"; remove SubType
)

type Calibration struct {
	device.Device `bson:",inline"`

	Units    *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value    *float64 `json:"value,omitempty" bson:"value,omitempty"`
	RawUnits *string  `json:"rawUnits,omitempty" bson:"rawUnits,omitempty"`
	RawValue *float64 `json:"rawValue,omitempty" bson:"rawValue,omitempty"`
}

func New() *Calibration {
	return &Calibration{
		Device: device.New(SubType),
	}
}

func (c *Calibration) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(c.Meta())
	}

	c.Device.Parse(parser)

	c.Units = parser.String("units")
	c.Value = parser.Float64("value")
}

func (c *Calibration) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(c.Meta())
	}

	c.Device.Validate(validator)

	if c.SubType != "" {
		validator.String("subType", &c.SubType).EqualTo(SubType)
	}

	validator.String("units", c.Units).Exists().OneOf(dataBloodGlucose.Units()...)
	validator.Float64("value", c.Value).Exists().InRange(dataBloodGlucose.ValueRangeForUnits(c.Units))
}

func (c *Calibration) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(c.Meta())
	}

	c.Device.Normalize(normalizer)

	if normalizer.Origin() == structure.OriginExternal {

		c.RawUnits = pointer.CloneString(c.Units)
		c.RawValue = pointer.CloneFloat64(c.Value)

		c.Units = dataBloodGlucose.NormalizeUnits(c.RawUnits)
		c.Value = dataBloodGlucose.NormalizeValueForUnits(c.RawValue, c.RawUnits)

	}
}
