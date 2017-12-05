package calibration

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/device"
)

type Calibration struct {
	device.Device `bson:",inline"`

	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func SubType() string {
	return "calibration"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Calibration {
	return &Calibration{}
}

func Init() *Calibration {
	calibration := New()
	calibration.Init()
	return calibration
}

func (c *Calibration) Init() {
	c.Device.Init()
	c.SubType = SubType()

	c.Value = nil
	c.Units = nil
}

func (c *Calibration) Parse(parser data.ObjectParser) error {
	if err := c.Device.Parse(parser); err != nil {
		return err
	}

	c.Value = parser.ParseFloat("value")
	c.Units = parser.ParseString("units")

	return nil
}

func (c *Calibration) Validate(validator data.Validator) error {
	if err := c.Device.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("subType", &c.SubType).EqualTo(SubType())

	validator.ValidateString("units", c.Units).Exists().OneOf(glucose.Units())
	validator.ValidateFloat("value", c.Value).Exists().InRange(glucose.ValueRangeForUnits(c.Units))

	return nil
}

func (c *Calibration) Normalize(normalizer data.Normalizer) {
	normalizer = normalizer.WithMeta(c.Meta())

	c.Device.Normalize(normalizer)

	c.Value = glucose.NormalizeValueForUnits(c.Value, c.Units)
	c.Units = glucose.NormalizeUnits(c.Units)
}
