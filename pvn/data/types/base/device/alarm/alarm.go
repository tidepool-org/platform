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
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/device"
)

type Alarm struct {
	device.Device
	AlarmType *string `json:"alarmType" bson:"alarmType"`
	Status    *string `json:"status,omitempty" bson:"status,omitempty"`
}

func Type() string {
	return device.Type()
}

func SubType() string {
	return "alarm"
}

func New() *Alarm {
	alarmType := Type()
	alarmSubType := SubType()

	alarm := &Alarm{}
	alarm.Type = &alarmType
	alarm.SubType = &alarmSubType
	return alarm
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

func (a *Alarm) Normalize(normalizer data.Normalizer) {
	a.Device.Normalize(normalizer)
}
