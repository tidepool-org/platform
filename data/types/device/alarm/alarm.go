package alarm

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/status"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	SubType = "alarm" // TODO: Rename Type to "device/alarm"; remove SubType

	AlarmTypeAutoOff    = "auto_off"
	AlarmTypeLowInsulin = "low_insulin"
	AlarmTypeLowPower   = "low_power"
	AlarmTypeNoDelivery = "no_delivery"
	AlarmTypeNoInsulin  = "no_insulin"
	AlarmTypeNoPower    = "no_power"
	AlarmTypeOcclusion  = "occlusion"
	AlarmTypeOther      = "other"
	AlarmTypeOverLimit  = "over_limit"
)

func AlarmTypes() []string {
	return []string{
		AlarmTypeAutoOff,
		AlarmTypeLowInsulin,
		AlarmTypeLowPower,
		AlarmTypeNoDelivery,
		AlarmTypeNoInsulin,
		AlarmTypeNoPower,
		AlarmTypeOcclusion,
		AlarmTypeOther,
		AlarmTypeOverLimit,
	}
}

type Alarm struct {
	device.Device `bson:",inline"`

	AlarmType *string        `json:"alarmType,omitempty" bson:"alarmType,omitempty"`
	Status    *status.Status `json:"-" bson:"-"`
	StatusID  *string        `json:"status,omitempty" bson:"status,omitempty"`
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
	a.SubType = SubType

	a.AlarmType = nil
	a.Status = nil
	a.StatusID = nil
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
		} else if *statusType != device.Type {
			statusParser.AppendError("type", service.ErrorValueStringNotOneOf(*statusType, []string{device.Type}))
		} else if statusSubType := statusParser.ParseString("subType"); statusSubType == nil {
			statusParser.AppendError("subType", service.ErrorValueNotExists())
		} else if *statusSubType != status.SubType {
			statusParser.AppendError("subType", service.ErrorValueStringNotOneOf(*statusSubType, []string{status.SubType}))
		} else if datum := parser.ParseDatum("status"); datum != nil {
			a.Status = (*datum).(*status.Status)
		}
	}

	return nil
}

func (a *Alarm) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(a.Meta())
	}

	a.Device.Validate(validator)

	if a.SubType != "" {
		validator.String("subType", &a.SubType).EqualTo(SubType)
	}

	validator.String("alarmType", a.AlarmType).Exists().OneOf(AlarmTypes()...)

	if validator.Origin() == structure.OriginExternal {
		if a.Status != nil {
			a.Status.Validate(validator.WithReference("status"))
		}
		validator.String("statusId", a.StatusID).NotExists()
	} else {
		if a.Status != nil {
			validator.WithReference("status").ReportError(structureValidator.ErrorValueExists())
		}
		validator.String("statusId", a.StatusID).Using(id.Validate)
	}
}

func (a *Alarm) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(a.Meta())
	}

	a.Device.Normalize(normalizer)

	if a.Status != nil {
		a.Status.Normalize(normalizer.WithReference("status"))
	}

	if normalizer.Origin() == structure.OriginExternal {
		if a.Status != nil {
			normalizer.AddData(a.Status)
			a.StatusID = a.Status.ID
			a.Status = nil
		}
	}
}
