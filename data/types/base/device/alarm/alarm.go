package alarm

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

type Alarm struct {
	device.Device `bson:",inline"`

	AlarmType *string `json:"alarmType,omitempty" bson:"alarmType,omitempty"`
	Status    *string `json:"status,omitempty" bson:"status,omitempty"`
}

func SubType() string {
	return "alarm"
}

func New() (*Alarm, error) {
	alarmDevice, err := device.New(SubType())
	if err != nil {
		return nil, err
	}

	return &Alarm{
		Device: *alarmDevice,
	}, nil
}

func (a *Alarm) Parse(parser data.ObjectParser) {
	a.Device.Parse(parser)

	a.AlarmType = parser.ParseString("alarmType")
	a.Status = parser.ParseString("status")
}

func (a *Alarm) Validate(validator data.Validator) {
	a.Device.Validate(validator)

	validator.ValidateString("alarmType", a.AlarmType).Exists().OneOf(
		[]string{
			"low_insulin",
			"no_insulin",
			"low_power",
			"no_power",
			"occlusion",
			"no_delivery",
			"auto_off",
			"over_limit",
			"other",
		},
	)

	validator.ValidateString("status", a.Status).LengthGreaterThan(1)
}
