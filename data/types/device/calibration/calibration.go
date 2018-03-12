package calibration

import (
	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = "calibration" // TODO: Rename Type to "device/calibration"; remove SubType
)

type Calibration struct {
	device.Device `bson:",inline"`

	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func New() *Calibration {
	return &Calibration{
		Device: device.New(SubType),
	}
}

func (c *Calibration) Parse(parser data.ObjectParser) error {
	if err := c.Device.Parse(parser); err != nil {
		return err
	}

	c.Units = parser.ParseString("units")
	c.Value = parser.ParseFloat("value")

	return nil
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
		units := c.Units
		c.Units = dataBloodGlucose.NormalizeUnits(units)
		c.Value = dataBloodGlucose.NormalizeValueForUnits(c.Value, units)
	}
}
