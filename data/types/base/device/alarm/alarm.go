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

func NewDatum() data.Datum {
	return New()
}

func New() *Alarm {
	return &Alarm{}
}

func Init() *Alarm {
	alarm := New()
	alarm.Init()
	return alarm
}

func (a *Alarm) Init() {
	a.Device.Init()
	a.Device.SubType = SubType()

	a.AlarmType = nil
	a.Status = nil
}

func (a *Alarm) Parse(parser data.ObjectParser) error {
	if err := a.Device.Parse(parser); err != nil {
		return err
	}

	a.AlarmType = parser.ParseString("alarmType")
	a.Status = parser.ParseString("status")

	return nil
}

func (a *Alarm) Validate(validator data.Validator) error {
	if err := a.Device.Validate(validator); err != nil {
		return err
	}

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

	return nil
}
