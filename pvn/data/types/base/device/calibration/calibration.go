package calibration

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/device"
)

type Calibration struct {
	device.Device

	Value *float64 `json:"value" bson:"value"`
	Units *string  `json:"units" bson:"units"`
}

func Type() string {
	return device.Type()
}

func SubType() string {
	return "calibration"
}

func New() *Calibration {
	calibrationType := Type()
	calibrationSubType := SubType()

	calibration := &Calibration{}
	calibration.Type = &calibrationType
	calibration.SubType = &calibrationSubType
	return calibration
}

func (c *Calibration) Parse(parser data.ObjectParser) {
	c.Device.Parse(parser)
	c.Value = parser.ParseFloat("value")
	c.Units = parser.ParseString("units")
}

func (c *Calibration) Validate(validator data.Validator) {
	c.Device.Validate(validator)
	validator.ValidateString("units", c.Units).Exists().OneOf([]string{"mmol/l", "mmol/L", "mg/dl", "mg/dL"})
	validator.ValidateFloat("value", c.Value).Exists().GreaterThanOrEqualTo(0.0).LessThanOrEqualTo(1000.0)
}

func (c *Calibration) Normalize(normalizer data.Normalizer) {
	c.Device.Normalize(normalizer)
}
