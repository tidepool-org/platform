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
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base/device"
)

type Calibration struct {
	device.Device `bson:",inline"`

	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func SubType() string {
	return "calibration"
}

func New() (*Calibration, error) {
	calibrationDevice, err := device.New(SubType())
	if err != nil {
		return nil, err
	}

	return &Calibration{
		Device: *calibrationDevice,
	}, nil
}

func (c *Calibration) Parse(parser data.ObjectParser) {
	c.Device.Parse(parser)

	c.Value = parser.ParseFloat("value")
	c.Units = parser.ParseString("units")
}

func (c *Calibration) Validate(validator data.Validator) {
	c.Device.Validate(validator)

	validator.ValidateStringAsBloodGlucoseUnits("units", c.Units).Exists()
	validator.ValidateFloatAsBloodGlucoseValue("value", c.Value).Exists().InRangeForUnits(c.Units)
}

func (c *Calibration) Normalize(normalizer data.Normalizer) {
	c.Device.Normalize(normalizer)

	c.Units, c.Value = normalizer.NormalizeBloodGlucose(c.Units).UnitsAndValue(c.Value)
}
