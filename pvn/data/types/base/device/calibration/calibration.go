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
	"github.com/tidepool-org/platform/pvn/data/types/common/bloodglucose"
)

type Calibration struct {
	device.Device `bson:",inline"`

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

	validator.ValidateString("units", c.Units).Exists().OneOf([]string{common.Mmoll, common.MmolL, common.Mgdl, common.MgdL})
	switch c.Units {
	case &common.Mmoll, &common.MmolL:
		validator.ValidateFloat("value", c.Value).Exists().InRange(common.MmolLFromValue, common.MmolLToValue)
	default:
		validator.ValidateFloat("value", c.Value).Exists().InRange(common.MgdLFromValue, common.MgdLToValue)
	}
}

func (c *Calibration) Normalize(normalizer data.Normalizer) {
	c.Device.Normalize(normalizer)

	c.Units, c.Value = normalizer.NormalizeBloodGlucose(Type(), c.Units).NormalizeUnitsAndValue(c.Value)
}
