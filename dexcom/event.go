package dexcom

import (
	"strconv"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypesActivityPhysical "github.com/tidepool-org/platform/data/types/activity/physical"
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	dataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	EventsResponseRecordType    = "event"
	EventsResponseRecordVersion = "3.0"

	EventStatusCreated = "created"
	EventStatusUpdated = "updated"
	EventStatusDeleted = "deleted"

	EventTypeUnknown        = "unknown"
	EventTypeInsulin        = "insulin"
	EventTypeCarbs          = "carbs"
	EventTypeExercise       = "exercise"
	EventTypeHealth         = "health"
	EventTypeBloodGlucose   = "bloodGlucose"
	EventTypeNotes          = "notes"
	EventTypeNotesAlternate = "note"

	EventSubTypeInsulinFastActing = "fastActing"
	EventSubTypeInsulinLongActing = "longActing"
	EventUnitInsulinUnits         = "units"
	EventValueInsulinUnitsMaximum = dataTypesInsulin.DoseTotalUnitsMaximum
	EventValueInsulinUnitsMinimum = dataTypesInsulin.DoseTotalUnitsMinimum
	EventValueInsulinUnitsDefault = "0"

	EventUnitCarbsGrams         = "grams"
	EventValueCarbsGramsMaximum = dataTypesFood.CarbohydrateNetGramsMaximum
	EventValueCarbsGramsMinimum = dataTypesFood.CarbohydrateNetGramsMinimum
	EventValueCarbsGramsDefault = "0"

	EventSubTypeExerciseLight        = "light"
	EventSubTypeExerciseMedium       = "medium"
	EventSubTypeExerciseHeavy        = "heavy"
	EventUnitExerciseMinutes         = "minutes"
	EventValueExerciseMinutesMaximum = dataTypesActivityPhysical.DurationValueMinutesMaximum
	EventValueExerciseMinutesMinimum = dataTypesActivityPhysical.DurationValueMinutesMinimum
	EventValueExerciseMinutesDefault = "0"

	EventSubTypeHealthIllness      = "illness"
	EventSubTypeHealthStress       = "stress"
	EventSubTypeHealthHighSymptoms = "highSymptoms"
	EventSubTypeHealthLowSymptoms  = "lowSymptoms"
	EventSubTypeHealthCycle        = "cycle"
	EventSubTypeHealthAlcohol      = "alcohol"

	EventUnitBloodGlucoseMgdL          = dataBloodGlucose.MgdL
	EventUnitBloodGlucoseMmolL         = dataBloodGlucose.MmolL
	EventValueBloodGlucoseMgdLMaximum  = dataBloodGlucose.MgdLMaximum
	EventValueBloodGlucoseMgdLMinimum  = dataBloodGlucose.MgdLMinimum
	EventValueBloodGlucoseMmolLMaximum = dataBloodGlucose.MmolLMaximum
	EventValueBloodGlucoseMmolLMinimum = dataBloodGlucose.MmolLMinimum
)

func EventStatuses() []string {
	return []string{
		EventStatusCreated,
		EventStatusUpdated,
		EventStatusDeleted,
	}
}

func EventTypes() []string {
	return []string{
		EventTypeUnknown,
		EventTypeInsulin,
		EventTypeCarbs,
		EventTypeExercise,
		EventTypeHealth,
		EventTypeBloodGlucose,
		EventTypeNotes,
	}
}

func EventSubTypesInsulin() []string {
	return []string{
		EventSubTypeInsulinFastActing,
		EventSubTypeInsulinLongActing,
	}
}

func EventUnitsInsulin() []string {
	return []string{
		EventUnitInsulinUnits,
	}
}

func EventUnitsCarbs() []string {
	return []string{
		EventUnitCarbsGrams,
	}
}

func EventSubTypesExercise() []string {
	return []string{
		EventSubTypeExerciseLight,
		EventSubTypeExerciseMedium,
		EventSubTypeExerciseHeavy,
	}
}

func EventUnitsExercise() []string {
	return []string{
		EventUnitExerciseMinutes,
	}
}

func EventSubTypesHealth() []string {
	return []string{
		EventSubTypeHealthIllness,
		EventSubTypeHealthStress,
		EventSubTypeHealthHighSymptoms,
		EventSubTypeHealthLowSymptoms,
		EventSubTypeHealthCycle,
		EventSubTypeHealthAlcohol,
	}
}

func EventUnitsBloodGlucose() []string {
	return []string{
		EventUnitBloodGlucoseMgdL,
		EventUnitBloodGlucoseMmolL,
	}
}

type EventsResponse struct {
	RecordType    *string `json:"recordType,omitempty"`
	RecordVersion *string `json:"recordVersion,omitempty"`
	UserID        *string `json:"userId,omitempty"`
	Records       *Events `json:"records,omitempty"`
}

func ParseEventsResponse(parser structure.ObjectParser) *EventsResponse {
	if !parser.Exists() {
		return nil
	}
	datum := NewEventsResponse()
	parser.Parse(datum)
	return datum
}

func NewEventsResponse() *EventsResponse {
	return &EventsResponse{}
}

func (e *EventsResponse) Parse(parser structure.ObjectParser) {
	parser = parser.WithMeta(e)

	e.RecordType = parser.String("recordType")
	e.RecordVersion = parser.String("recordVersion")
	e.UserID = parser.String("userId")
	e.Records = ParseEvents(parser.WithReferenceArrayParser("records"))
}

func (e *EventsResponse) Validate(validator structure.Validator) {
	validator = validator.WithMeta(e)

	validator.String("recordType", e.RecordType).Exists().EqualTo(EventsResponseRecordType)
	validator.String("recordVersion", e.RecordVersion).Exists().EqualTo(EventsResponseRecordVersion)
	validator.String("userId", e.UserID).Exists().NotEmpty()

	// Only validate that the records exist, remaining validation will occur later on a per-record basis
	if e.Records == nil {
		validator.WithReference("records").ReportError(structureValidator.ErrorValueNotExists())
	}
}

type Events []*Event

func ParseEvents(parser structure.ArrayParser) *Events {
	if !parser.Exists() {
		return nil
	}
	datum := NewEvents()
	parser.Parse(datum)
	return datum
}

func NewEvents() *Events {
	return &Events{}
}

func (e *Events) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*e = append(*e, ParseEvent(parser.WithReferenceObjectParser(reference)))
	}
}

type Event struct {
	RecordID              *string `json:"recordId,omitempty"`
	SystemTime            *Time   `json:"systemTime,omitempty"`
	DisplayTime           *Time   `json:"displayTime,omitempty"`
	EventStatus           *string `json:"eventStatus,omitempty"`
	EventType             *string `json:"eventType,omitempty"`
	EventSubType          *string `json:"eventSubType,omitempty"`
	Unit                  *string `json:"unit,omitempty"`
	Value                 *string `json:"value,omitempty"`
	TransmitterGeneration *string `json:"transmitterGeneration,omitempty"`
	TransmitterID         *string `json:"transmitterId,omitempty"`
	DisplayDevice         *string `json:"displayDevice,omitempty"`
}

func ParseEvent(parser structure.ObjectParser) *Event {
	if !parser.Exists() {
		return nil
	}
	datum := NewEvent()
	parser.Parse(datum)
	return datum
}

func NewEvent() *Event {
	return &Event{}
}

func (e *Event) Parse(parser structure.ObjectParser) {
	parser = parser.WithMeta(e)

	e.RecordID = parser.String("recordId")
	e.SystemTime = ParseTime(parser, "systemTime")
	e.DisplayTime = ParseTime(parser, "displayTime")
	e.EventStatus = parser.String("eventStatus")
	e.EventType = parser.String("eventType")
	e.EventSubType = parser.String("eventSubType")
	e.Unit = parser.String("unit")
	e.Value = parser.String("value")
	e.TransmitterGeneration = parser.String("transmitterGeneration")
	e.TransmitterID = parser.String("transmitterId")
	e.DisplayDevice = parser.String("displayDevice")
}

func (e *Event) Validate(validator structure.Validator) {
	validator = validator.WithMeta(e)

	if e.EventType != nil && *e.EventType == EventTypeNotesAlternate {
		validator.Logger().Warnf("EventType is '%s'", *e.EventType)
		e.EventType = pointer.FromString(EventTypeNotes)
	}

	validator.String("recordId", e.RecordID).Exists().NotEmpty()
	validator.Time("systemTime", e.SystemTime.Raw()).Exists().NotZero()
	validator.Time("displayTime", e.DisplayTime.Raw()).Exists().NotZero()
	validator.String("eventStatus", e.EventStatus).Exists().OneOf(EventStatuses()...)
	validator.String("eventType", e.EventType).Exists().OneOf(EventTypes()...)
	if e.EventType != nil {
		switch *e.EventType {
		case EventTypeUnknown:
			e.validateUnknown(validator)
		case EventTypeInsulin:
			e.validateInsulin(validator)
		case EventTypeCarbs:
			e.validateCarbs(validator)
		case EventTypeExercise:
			e.validateExercise(validator)
		case EventTypeHealth:
			e.validateHealth(validator)
		case EventTypeBloodGlucose:
			e.validateBloodGlucose(validator)
		case EventTypeNotes:
			e.validateNote(validator)
		}
	}
	validator.String("transmitterGeneration", e.TransmitterGeneration).Exists().OneOf(DeviceTransmitterGenerations()...)
	validator.String("transmitterId", e.TransmitterID).Exists().Using(TransmitterIDValidator)
	validator.String("displayDevice", e.DisplayDevice).Exists().OneOf(DeviceDisplayDevices()...)

	// Log various warnings
	logger := validator.Logger().WithField("meta", e)
	if e.EventStatus != nil && *e.EventStatus != EventStatusCreated {
		logger.Warnf("EventStatus is '%s'", *e.EventStatus)
	}
	if e.EventType != nil && *e.EventType == EventTypeUnknown {
		logger.Warnf("EventType is '%s'", *e.EventType)
	}
	if e.TransmitterID != nil && *e.TransmitterID == "" {
		logger.Warnf("TransmitterID is empty", *e.TransmitterID)
	}
	if e.DisplayDevice != nil && *e.DisplayDevice == DeviceDisplayDeviceUnknown {
		logger.Warnf("DisplayDevice is '%s'", *e.DisplayDevice)
	}
}

func (e *Event) Normalize(normalizer structure.Normalizer) {}

func (e *Event) validateUnknown(validator structure.Validator) {}

func (e *Event) validateInsulin(validator structure.Validator) {

	// HACK: Dexcom - Fix missing or invalid unit or value
	if e.Unit == nil || *e.Unit == "" {
		validator.Logger().Warnf("Event has missing or invalid unit")
		e.Unit = pointer.FromString(EventUnitInsulinUnits)
	}
	if e.Value == nil || *e.Value == "" {
		validator.Logger().Warnf("Event has missing or invalid value")
		e.Value = pointer.FromString(EventValueInsulinUnitsDefault)
	}

	validator.String("eventSubType", e.EventSubType).OneOf(EventSubTypesInsulin()...) // Dexcom - May not exist
	validator.String("unit", e.Unit).Exists().OneOf(EventUnitsInsulin()...)
	validator.String("value", e.Value).Exists().NotEmpty()
	if e.Unit != nil {
		switch *e.Unit {
		case EventUnitInsulinUnits:
			validateValueAsFloat64AndInRange(validator, e.Value, EventValueInsulinUnitsMinimum, EventValueInsulinUnitsMaximum)
		default:
			validateValueAsFloat64(validator, e.Value)
		}
	}
}

func (e *Event) validateCarbs(validator structure.Validator) {

	// HACK: Dexcom - Fix missing or invalid unit or value
	if e.Unit == nil || *e.Unit == "" {
		validator.Logger().Warnf("Event has missing or invalid unit")
		e.Unit = pointer.FromString(EventUnitCarbsGrams)
	}
	if e.Value == nil || *e.Value == "" {
		validator.Logger().Warnf("Event has missing or invalid value")
		e.Value = pointer.FromString(EventValueCarbsGramsDefault)
	}

	validator.String("eventSubType", e.EventSubType).NotExists()
	validator.String("unit", e.Unit).Exists().OneOf(EventUnitsCarbs()...)
	validator.String("value", e.Value).Exists().NotEmpty()
	if e.Unit != nil {
		switch *e.Unit {
		case EventUnitCarbsGrams:
			validateValueAsFloat64AndInRange(validator, e.Value, EventValueCarbsGramsMinimum, EventValueCarbsGramsMaximum)
		default:
			validateValueAsFloat64(validator, e.Value)
		}
	}
}

func (e *Event) validateExercise(validator structure.Validator) {

	// HACK: Dexcom - Fix missing or invalid unit or value
	if e.Unit == nil || *e.Unit == "" {
		validator.Logger().Warnf("Event has missing or invalid unit")
		e.Unit = pointer.FromString(EventUnitExerciseMinutes)
	}
	if e.Value == nil || *e.Value == "" || *e.Value == "-1" || *e.Value == "-1.00" {
		validator.Logger().Warnf("Event has missing or invalid value")
		e.Value = pointer.FromString(EventValueExerciseMinutesDefault)
	}

	validator.String("eventSubType", e.EventSubType).OneOf(EventSubTypesExercise()...) // Dexcom - May not exist
	validator.String("unit", e.Unit).Exists().OneOf(EventUnitsExercise()...)
	validator.String("value", e.Value).Exists().NotEmpty()
	if e.Unit != nil {
		switch *e.Unit {
		case EventUnitExerciseMinutes:
			validateValueAsFloat64AndInRange(validator, e.Value, EventValueExerciseMinutesMinimum, EventValueExerciseMinutesMaximum)
		default:
			validateValueAsFloat64(validator, e.Value)
		}
	}
}

func (e *Event) validateHealth(validator structure.Validator) {
	if e.Value != nil {
		validator.Logger().Warnf("Event has invalid value")
		e.Value = nil
	}

	validator.String("eventSubType", e.EventSubType).Exists().OneOf(EventSubTypesHealth()...)
	validator.String("unit", e.Unit).NotExists()
	validator.String("value", e.Value).NotExists()
}

func (e *Event) validateBloodGlucose(validator structure.Validator) {
	validator.String("eventSubType", e.EventSubType).NotExists()
	validator.String("unit", e.Unit).Exists().OneOf(EventUnitsBloodGlucose()...)
	validator.String("value", e.Value).Exists().NotEmpty()
	if e.Unit != nil {
		switch *e.Unit {
		case EventUnitBloodGlucoseMgdL:
			validateValueAsFloat64AndInRange(validator, e.Value, EventValueBloodGlucoseMgdLMinimum, EventValueBloodGlucoseMgdLMaximum)
		case EventUnitBloodGlucoseMmolL:
			validateValueAsFloat64AndInRange(validator, e.Value, EventValueBloodGlucoseMmolLMinimum, EventValueBloodGlucoseMmolLMaximum)
		default:
			validateValueAsFloat64(validator, e.Value)
		}
	}
}

func (e *Event) validateNote(validator structure.Validator) {
	validator.String("eventSubType", e.EventSubType).NotExists()
	validator.String("unit", e.Unit).NotExists()
	validator.String("value", e.Value).Exists().NotEmpty()
}

func validateValueAsFloat64(validator structure.Validator, value *string) {
	if value != nil {
		if _, err := strconv.ParseFloat(*value, 64); err != nil {
			validator.WithReference("value").ReportError(ErrorValueFloat64NotParsable(*value))
		}
	}
}

func validateValueAsFloat64AndInRange(validator structure.Validator, value *string, lowerLimit float64, upperLimit float64) {
	if value != nil {
		if floatVal, err := strconv.ParseFloat(*value, 64); err != nil {
			validator.WithReference("value").ReportError(ErrorValueFloat64NotParsable(*value))
		} else {
			validator.Float64("value", &floatVal).InRange(lowerLimit, upperLimit)
		}
	}
}

func ErrorValueFloat64NotParsable(value string) error {
	return errors.Preparedf(structureParser.ErrorCodeValueNotParsable, "value is not a parsable float64", "value %q is not a parsable float64", value)
}
