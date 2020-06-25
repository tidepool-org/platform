package alarm

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/device"
	dataTypesDeviceStatus "github.com/tidepool-org/platform/data/types/device/status"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	SubType = "alarm" // TODO: Rename Type to "device/alarm"; remove SubType

	AlarmTypeAutoOff        = "auto_off"
	AlarmTypeLowInsulin     = "low_insulin"
	AlarmTypeLowPower       = "low_power"
	AlarmTypeNoDelivery     = "no_delivery"
	AlarmTypeNoInsulin      = "no_insulin"
	AlarmTypeNoPower        = "no_power"
	AlarmTypeOcclusion      = "occlusion"
	AlarmTypeOther          = "other"
	AlarmTypeOverLimit      = "over_limit"
	AlarmTypeHandset        = "handset"
	IsAnAlarm               = "alarm"
	IsAnAlert               = "alert"
	NewAck                  = "new"
	Acknowledged            = "acknowledged"
	Outdated                = "outdated"
	AlarmCodeMaximumLength  = 64
	EventIDMaximumLength    = 64
	AlarmLabelMaximumLength = 256
)

func LegacyAlarmTypes() []string {
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

func AlarmTypes() []string {
	return append(LegacyAlarmTypes(), AlarmTypeHandset)
}

func AlarmLevels() []string {
	return []string{
		IsAnAlarm,
		IsAnAlert,
	}
}

func AckStatuses() []string {
	return []string{
		NewAck,
		Acknowledged,
		Outdated,
	}
}

type Alarm struct {
	device.Device `bson:",inline"`

	AlarmType  *string     `json:"alarmType,omitempty" bson:"alarmType,omitempty"`
	Status     *data.Datum `json:"-" bson:"-"`
	StatusID   *string     `json:"status,omitempty" bson:"status,omitempty"`
	EventID    *string     `json:"eventId,omitempty" bson:"eventId,omitempty"`
	AlarmLevel *string     `json:"alarmLevel,omitempty" bson:"alarmLevel,omitempty"`
	AlarmCode  *string     `json:"alarmCode,omitempty" bson:"alarmCode,omitempty"`
	AlarmLabel *string     `json:"alarmLabel,omitempty" bson:"alarmLabel,omitempty"`
	AckStatus  *string     `json:"ackStatus,omitempty" bson:"ackStatus,omitempty"`
	UpdateTime *string     `json:"updateTime,omitempty" bson:"updateTime,omitempty"`
}

func New() *Alarm {
	return &Alarm{
		Device: device.New(SubType),
	}
}

func (a *Alarm) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(a.Meta())
	}

	a.Device.Parse(parser)

	a.AlarmType = parser.String("alarmType")
	a.Status = dataTypesDeviceStatus.ParseStatusDatum(parser.WithReferenceObjectParser("status"))
	a.EventID = parser.String("eventId")
	a.AlarmLevel = parser.String("alarmLevel")
	a.AlarmCode = parser.String("alarmCode")
	a.AlarmLabel = parser.String("alarmLabel")
	a.AckStatus = parser.String("ackStatus")
	a.UpdateTime = parser.String("updateTime")
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
			(*a.Status).Validate(validator.WithReference("status"))
		}
		validator.String("statusId", a.StatusID).NotExists()
	} else {
		if a.Status != nil {
			validator.WithReference("status").ReportError(structureValidator.ErrorValueExists())
		}
		validator.String("statusId", a.StatusID).Using(data.IDValidator)
	}

	alarmLevelValidator := validator.String("alarmLevel", a.AlarmLevel)
	alarmLevelValidator.OneOf(AlarmLevels()...)

	ackStatusValidator := validator.String("ackStatus", a.AckStatus)
	ackStatusValidator.OneOf(AckStatuses()...)

	timeValidator := validator.String("updateTime", a.UpdateTime)
	timeValidator.AsTime(types.TimeFormat)

	if a.AlarmType != nil && *a.AlarmType == AlarmTypeHandset {
		validator.String("eventID", a.EventID).Exists().LengthLessThanOrEqualTo(EventIDMaximumLength)
		alarmLevelValidator.Exists()
		validator.String("alarmCode", a.AlarmCode).Exists().LengthLessThanOrEqualTo(AlarmCodeMaximumLength)
		validator.String("alarmLabel", a.AlarmLabel).Exists().LengthLessThanOrEqualTo(AlarmLabelMaximumLength)
		ackStatusValidator.Exists()
		timeValidator.Exists()
	}
}

// IsValid returns true if there is no error in the validator
func (a *Alarm) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (a *Alarm) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(a.Meta())
	}

	a.Device.Normalize(normalizer)

	if a.Status != nil {
		(*a.Status).Normalize(normalizer.WithReference("status"))
	}

	if normalizer.Origin() == structure.OriginExternal {
		if a.Status != nil {
			normalizer.AddData(*a.Status)
			switch status := (*a.Status).(type) {
			case *dataTypesDeviceStatus.Status:
				a.StatusID = status.ID
			}
			a.Status = nil
		}
	}
}
