package data

import (
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	getPlatformValidator().RegisterValidation(deviceEventSubTypeField.Tag, DeviceEventSubTypeValidator)
	getPlatformValidator().RegisterValidation(deviceEventAlarmTypeField.Tag, DeviceEventAlarmTypeValidator)
	getPlatformValidator().RegisterValidation(deviceEventUnitsField.Tag, DeviceEventUnitsValidator)
	getPlatformValidator().RegisterValidation(deviceEventValueField.Tag, DeviceEventValueValidator)
	getPlatformValidator().RegisterValidation(deviceEventStatusField.Tag, DeviceEventStatusValidator)
	getPlatformValidator().RegisterValidation(deviceEventReasonField.Tag, DeviceEventReasonValidator)
	getPlatformValidator().RegisterValidation(deviceEventPrimeTargetField.Tag, DeviceEventPrimeTargetValidator)
	getPlatformValidator().RegisterValidation(deviceEventVolumeField.Tag, DeviceEventVolumeValidator)
	getPlatformValidator().RegisterValidation(deviceEventTimeChangeReasonsField.Tag, DeviceEventTimeChangeReasonsValidator)
	getPlatformValidator().RegisterValidation(deviceEventTimeChangeAgentField.Tag, DeviceEventTimeChangeAgentValidator)
}

type DeviceEvent struct {
	SubType *string `json:"subType" bson:"subType" valid:"deviceeventsubtype"`
	Base    `bson:",inline"`
}

type AlarmDeviceEvent struct {
	AlarmType   *string `json:"alarmType" bson:"alarmType" valid:"deviceeventalarmtype"`
	DeviceEvent `bson:",inline"`
}

type CalibrationDeviceEvent struct {
	Value       *float64 `json:"value" bson:"value" valid:"deviceeventvalue"`
	Units       *string  `json:"units" bson:"units" valid:"deviceeventunits"`
	DeviceEvent `bson:",inline"`
}

type StatusDeviceEvent struct {
	Status      *string      `json:"status" bson:"status" valid:"deviceeventstatus"`
	Reason      *interface{} `json:"reason" bson:"reason" valid:"deviceeventreason"`
	DeviceEvent `bson:",inline"`
}

type PrimeDeviceEvent struct {
	PrimeTarget *string  `json:"primeTarget" bson:"primeTarget" valid:"deviceeventprimetarget"`
	Volume      *float64 `json:"volume,omitempty" bson:"volume,omitempty" valid:"omitempty,deviceeventvolume"`
	DeviceEvent `bson:",inline"`
}

type TimeChangeDeviceEvent struct {
	Status      *string `json:"status" bson:"status" valid:"deviceeventstatus"`
	Change      `json:"change" bson:"change"`
	DeviceEvent `bson:",inline"`
}

type Change struct {
	From     *string   `json:"from" bson:"from" valid:"timestr"`
	To       *string   `json:"to" bson:"to" valid:"timestr"`
	Agent    *string   `json:"from" bson:"from" valid:"deviceeventchangeagent"`
	Timezone *string   `json:"timezone,omitempty" bson:"timezone,omitempty" valid:"-"`
	Reasons  *[]string `json:"reasons,omitempty" bson:"reasons,omitempty" valid:"omitempty,deviceeventchangereasons"`
}

type ReservoirChangeDeviceEvent struct {
	Status      *string `json:"status" bson:"status" valid:"deviceeventstatus"`
	DeviceEvent `bson:",inline"`
}

const DeviceEventName = "deviceEvent"

var (
	deviceEventSubTypeField = TypesDatumField{
		DatumField: &DatumField{Name: "subType"},
		Tag:        "deviceeventsubtype",
		Message:    "Must be one of alarm, calibration, status, prime, timeChange, reservoirChange",
		AllowedTypes: AllowedTypes{
			"alarm":           true,
			"calibration":     true,
			"status":          true,
			"prime":           true,
			"timeChange":      true,
			"reservoirChange": true,
		},
	}

	deviceEventUnitsField = TypesDatumField{
		DatumField: &DatumField{Name: "units"},
		Tag:        "deviceeventunits",
		Message:    "Must be one of mg/dl, mmol/l",
		AllowedTypes: AllowedTypes{
			"mmol/L": true,
			"mmol/l": true,
			"mg/dL":  true,
			"mg/dl":  true,
		},
	}

	deviceEventValueField = FloatDatumField{
		DatumField:        &DatumField{Name: "value"},
		Tag:               "deviceeventvalue",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &AllowedFloatRange{LowerLimit: 0.0},
	}

	deviceEventVolumeField = FloatDatumField{
		DatumField:        &DatumField{Name: "volume"},
		Tag:               "deviceeventvolume",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &AllowedFloatRange{LowerLimit: 0.0},
	}

	deviceEventAlarmTypeField = TypesDatumField{
		DatumField: &DatumField{Name: "alarmType"},
		Tag:        "deviceeventalarmtype",
		Message:    "Must be one of low_insulin, no_insulin, low_power, no_power, occlusion, no_delivery, auto_off, over_limit, other",
		AllowedTypes: AllowedTypes{
			"low_insulin": true,
			"no_insulin":  true,
			"low_power":   true,
			"no_power":    true,
			"occlusion":   true,
			"no_delivery": true,
			"auto_off":    true,
			"over_limit":  true,
			"other":       true,
		},
	}

	deviceEventReasonField = TypesDatumField{
		DatumField: &DatumField{Name: "reason"},
		Tag:        "deviceeventreason",
		Message:    "Must be one of manual, automatic",
		AllowedTypes: AllowedTypes{
			"manual":    true,
			"automatic": true,
		},
	}

	deviceEventStatusField = TypesDatumField{
		DatumField: &DatumField{Name: "status"},
		Tag:        "deviceeventstatus",
		Message:    "Must be one of suspended, resumed",
		AllowedTypes: AllowedTypes{
			"suspended": true,
			"resumed":   true,
		},
	}

	deviceEventPrimeTargetField = TypesDatumField{
		DatumField: &DatumField{Name: "primeTarget"},
		Tag:        "deviceeventprimetarget",
		Message:    "Must be one of cannula, tubing",
		AllowedTypes: AllowedTypes{
			"cannula": true,
			"tubing":  true,
		},
	}

	deviceEventTimeChangeReasonsField = TypesDatumField{
		DatumField: &DatumField{Name: "reasons"},
		Tag:        "deviceeventchangereasons",
		Message:    "Must be one of from_daylight_savings, to_daylight_savings, travel, correction, other",
		AllowedTypes: AllowedTypes{
			"from_daylight_savings": true,
			"to_daylight_savings":   true,
			"travel":                true,
			"correction":            true,
			"other":                 true,
		},
	}

	deviceEventTimeChangeAgentField = TypesDatumField{
		DatumField: &DatumField{Name: "agent"},
		Tag:        "deviceeventchangeagent",
		Message:    "Must be one of manual, automatic",
		AllowedTypes: AllowedTypes{
			"manual":    true,
			"automatic": true,
		},
	}

	deviceEventTimeChangeFromField     = DatumField{Name: "from"}
	deviceEventTimeChangeToField       = DatumField{Name: "to"}
	deviceEventTimeChangeTimezoneField = DatumField{Name: "timezone"}
)

func (de DeviceEvent) makeAlarm(datum Datum, errs validate.ErrorProcessing) *AlarmDeviceEvent {
	alarmDeviceEvent := &AlarmDeviceEvent{
		AlarmType:   datum.ToString(deviceEventAlarmTypeField.Name, errs),
		DeviceEvent: de,
	}
	getPlatformValidator().Struct(alarmDeviceEvent, errs)
	return alarmDeviceEvent
}

func (de DeviceEvent) makeCalibration(datum Datum, errs validate.ErrorProcessing) *CalibrationDeviceEvent {
	calibrationDeviceEvent := &CalibrationDeviceEvent{
		Value:       datum.ToFloat64(deviceEventValueField.Name, errs),
		Units:       datum.ToString(deviceEventUnitsField.Name, errs),
		DeviceEvent: de,
	}
	getPlatformValidator().Struct(calibrationDeviceEvent, errs)
	return calibrationDeviceEvent
}

func (de DeviceEvent) makePrime(datum Datum, errs validate.ErrorProcessing) *PrimeDeviceEvent {
	primeDeviceEvent := &PrimeDeviceEvent{
		PrimeTarget: datum.ToString(deviceEventPrimeTargetField.Name, errs),
		Volume:      datum.ToFloat64(deviceEventVolumeField.Name, errs),
		DeviceEvent: de,
	}
	getPlatformValidator().Struct(primeDeviceEvent, errs)
	return primeDeviceEvent
}

func (de DeviceEvent) makeStatus(datum Datum, errs validate.ErrorProcessing) *StatusDeviceEvent {
	statusDeviceEvent := &StatusDeviceEvent{
		Status:      datum.ToString(deviceEventStatusField.Name, errs),
		Reason:      datum.ToObject(deviceEventReasonField.Name, errs),
		DeviceEvent: de,
	}
	getPlatformValidator().Struct(statusDeviceEvent, errs)
	return statusDeviceEvent
}

func makeChange(datum Datum, errs validate.ErrorProcessing) Change {
	return Change{
		From:     datum.ToString(deviceEventTimeChangeFromField.Name, errs),
		To:       datum.ToString(deviceEventTimeChangeToField.Name, errs),
		Agent:    datum.ToString(deviceEventTimeChangeAgentField.Name, errs),
		Timezone: datum.ToString(deviceEventTimeChangeTimezoneField.Name, errs),
	}
}

func (de DeviceEvent) makeTimeChange(datum Datum, errs validate.ErrorProcessing) *TimeChangeDeviceEvent {

	timeChangeDeviceEvent := &TimeChangeDeviceEvent{
		Change:      makeChange(datum["change"].(map[string]interface{}), errs),
		DeviceEvent: de,
	}

	getPlatformValidator().Struct(timeChangeDeviceEvent, errs)
	return timeChangeDeviceEvent
}

func (de DeviceEvent) makeReservoirChange(datum Datum, errs validate.ErrorProcessing) *ReservoirChangeDeviceEvent {
	reservoirChangeDeviceEvent := &ReservoirChangeDeviceEvent{
		Status:      datum.ToString(deviceEventStatusField.Name, errs),
		DeviceEvent: de,
	}
	getPlatformValidator().Struct(reservoirChangeDeviceEvent, errs)
	return reservoirChangeDeviceEvent
}

func BuildDeviceEvent(datum Datum, errs validate.ErrorProcessing) interface{} {

	deviceEvent := DeviceEvent{
		SubType: datum.ToString(deviceEventSubTypeField.Name, errs),
		Base:    BuildBase(datum, errs),
	}

	switch *deviceEvent.SubType {
	case "alarm":
		return deviceEvent.makeAlarm(datum, errs)
	case "calibration":
		return deviceEvent.makeCalibration(datum, errs)
	case "status":
		return deviceEvent.makeStatus(datum, errs)
	case "prime":
		return deviceEvent.makePrime(datum, errs)
	case "timeChange":
		return deviceEvent.makeTimeChange(datum, errs)
	case "reservoirChange":
		return deviceEvent.makeReservoirChange(datum, errs)
	}
	return nil
}

func DeviceEventTimeChangeAgentValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	agent, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = deviceEventTimeChangeAgentField.AllowedTypes[agent]
	return ok
}

func DeviceEventTimeChangeReasonsValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	reason, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = deviceEventTimeChangeReasonsField.AllowedTypes[reason]
	return ok
}

func DeviceEventValueValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	val, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return val > deviceEventValueField.LowerLimit
}

func DeviceEventVolumeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	volume, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return volume > deviceEventVolumeField.LowerLimit
}

func DeviceEventPrimeTargetValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	target, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = deviceEventPrimeTargetField.AllowedTypes[target]
	return ok
}

func DeviceEventStatusValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	status, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = deviceEventStatusField.AllowedTypes[status]
	return ok
}

func DeviceEventReasonValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	reason, ok := field.Interface().(map[string]string)
	if !ok {
		return false
	}
	ok = false
	for _, val := range reason {
		_, ok = deviceEventReasonField.AllowedTypes[val]
	}
	return ok
}

func DeviceEventUnitsValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	units, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = deviceEventUnitsField.AllowedTypes[units]
	return ok
}

func DeviceEventSubTypeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	alarmType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = deviceEventAlarmTypeField.AllowedTypes[alarmType]
	return ok
}

func DeviceEventAlarmTypeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	subType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = deviceEventSubTypeField.AllowedTypes[subType]
	return ok
}
