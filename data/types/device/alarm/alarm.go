package alarm

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/status"
	"github.com/tidepool-org/platform/service"
)

type Alarm struct {
	device.Device `bson:",inline"`

	AlarmType *string `json:"alarmType,omitempty" bson:"alarmType,omitempty"`
	StatusID  *string `json:"status,omitempty" bson:"status,omitempty"`

	// Embedded status
	status *data.Datum
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
	a.SubType = SubType()

	a.AlarmType = nil
	a.StatusID = nil

	a.status = nil
}

func (a *Alarm) Parse(parser data.ObjectParser) error {
	if err := a.Device.Parse(parser); err != nil {
		return err
	}

	a.AlarmType = parser.ParseString("alarmType")

	// TODO: This is a bit hacky to ensure we only parse true status data. Is there a better way?

	if statusParser := parser.NewChildObjectParser("status"); statusParser.Object() != nil {
		if statusType := statusParser.ParseString("type"); statusType == nil {
			statusParser.AppendError("type", service.ErrorValueNotExists())
		} else if *statusType != device.Type() {
			statusParser.AppendError("type", service.ErrorValueStringNotOneOf(*statusType, []string{device.Type()}))
		} else if statusSubType := statusParser.ParseString("subType"); statusSubType == nil {
			statusParser.AppendError("subType", service.ErrorValueNotExists())
		} else if *statusSubType != status.SubType() {
			statusParser.AppendError("subType", service.ErrorValueStringNotOneOf(*statusSubType, []string{status.SubType()}))
		} else {
			a.status = parser.ParseDatum("status")
		}
	}

	return nil
}

func (a *Alarm) Validate(validator data.Validator) error {
	if err := a.Device.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("subType", &a.SubType).EqualTo(SubType())

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

	if a.status != nil {
		(*a.status).Validate(validator.NewChildValidator("status"))
	}

	return nil
}

func (a *Alarm) Normalize(normalizer data.Normalizer) {
	normalizer = normalizer.WithMeta(a.Meta())

	a.Device.Normalize(normalizer)

	if a.status != nil {
		(*a.status).Normalize(normalizer.WithReference("status"))

		a.StatusID = &(*a.status).(*status.Status).ID

		normalizer.AddData(*a.status)
		a.status = nil
	}
}
