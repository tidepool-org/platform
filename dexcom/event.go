package dexcom

import (
	"strconv"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypesActivityPhysical "github.com/tidepool-org/platform/data/types/activity/physical"
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	dataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	EventTypeCarbs    = "carbs"
	EventTypeExercise = "exercise"
	EventTypeHealth   = "health"
	EventTypeInsulin  = "insulin"
	EventTypeUnknown  = "unknown"
	EventTypeBG       = "bloodGlucose"
	EventTypeNote     = "note"
	EventTypeNotes    = "notes"

	EventUnitUnknown = "unknown"

	EventUnitMgdL         = dataBloodGlucose.MgdL
	EventValueMgdLMaximum = dataBloodGlucose.MgdLMaximum
	EventValueMgdLMinimum = dataBloodGlucose.MgdLMinimum

	EventUnitCarbsGrams         = "grams"
	EventValueCarbsGramsMaximum = dataTypesFood.CarbohydrateNetGramsMaximum
	EventValueCarbsGramsMinimum = dataTypesFood.CarbohydrateNetGramsMinimum

	EventSubTypeExerciseLight        = "light"
	EventSubTypeExerciseMedium       = "medium"
	EventSubTypeExerciseHeavy        = "heavy"
	EventUnitExerciseMinutes         = "minutes"
	EventValueExerciseMinutesMaximum = dataTypesActivityPhysical.DurationValueMinutesMaximum
	EventValueExerciseMinutesMinimum = dataTypesActivityPhysical.DurationValueMinutesMinimum

	EventSubTypeHealthAlcohol      = "alcohol"
	EventSubTypeHealthCycle        = "cycle"
	EventSubTypeHealthHighSymptoms = "highSymptoms"
	EventSubTypeHealthIllness      = "illness"
	EventSubTypeHealthLowSymptoms  = "lowSymptoms"
	EventSubTypeHealthStress       = "stress"

	EventSubTypeInsulinFastActing = "fastActing"
	EventSubTypeInsulinLongActing = "longActing"
	EventUnitInsulinUnits         = "units"
	EventValueInsulinUnitsMaximum = dataTypesInsulin.DoseTotalUnitsMaximum
	EventValueInsulinUnitsMinimum = dataTypesInsulin.DoseTotalUnitsMinimum

	EventStatusCreated = "created"
	EventStatusUpdated = "updated"
	EventStatusDeleted = "deleted"

	eventErrorValue = -100
)

func EventTypes() []string {
	return []string{
		EventTypeBG,
		EventTypeCarbs,
		EventTypeExercise,
		EventTypeHealth,
		EventTypeInsulin,
		EventTypeNote,
		EventTypeNotes,
		EventTypeUnknown,
	}
}

func EventSubTypesExercise() []string {
	return []string{
		EventSubTypeExerciseLight,
		EventSubTypeExerciseMedium,
		EventSubTypeExerciseHeavy,
	}
}

func EventSubTypesHealth() []string {
	return []string{
		EventSubTypeHealthAlcohol,
		EventSubTypeHealthCycle,
		EventSubTypeHealthHighSymptoms,
		EventSubTypeHealthIllness,
		EventSubTypeHealthLowSymptoms,
		EventSubTypeHealthStress,
	}
}

func EventSubTypesInsulin() []string {
	return []string{
		EventSubTypeInsulinFastActing,
		EventSubTypeInsulinLongActing,
	}
}

func EventStatuses() []string {
	return []string{
		EventStatusCreated,
		EventStatusUpdated,
		EventStatusDeleted,
	}
}

type EventsResponse struct {
	RecordType    *string `json:"recordType,omitempty"`
	RecordVersion *string `json:"recordVersion,omitempty"`
	UserID        *string `json:"userId,omitempty"`
	Events        *Events `json:"records,omitempty"`
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
	e.UserID = parser.String("userId")
	e.RecordType = parser.String("recordType")
	e.RecordVersion = parser.String("recordVersion")
	e.Events = ParseEvents(parser.WithReferenceArrayParser("records"))
}

func (e *EventsResponse) Validate(validator structure.Validator) {
	if eventsValidator := validator.WithReference("records"); e.Events != nil {
		e.Events.Validate(eventsValidator)
	} else {
		eventsValidator.ReportError(structureValidator.ErrorValueNotExists())
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

func (e *Events) Validate(validator structure.Validator) {
	for index, event := range *e {
		if eventValidator := validator.WithReference(strconv.Itoa(index)); event != nil {
			event.Validate(eventValidator)
		} else {
			eventValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type Event struct {
	ID                    *string `json:"recordId,omitempty"`
	SystemTime            *Time   `json:"systemTime,omitempty"`
	DisplayTime           *Time   `json:"displayTime,omitempty"`
	Type                  *string `json:"eventType,omitempty"`
	SubType               *string `json:"eventSubType,omitempty"`
	Unit                  *string `json:"unit,omitempty"`
	Value                 *string `json:"value,omitempty"`
	Status                *string `json:"eventStatus,omitempty"`
	TransmitterID         *string `json:"transmitterId,omitempty"`
	TransmitterGeneration *string `json:"transmitterGeneration,omitempty"`
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
	e.SystemTime = ParseTime(parser, "systemTime")
	e.DisplayTime = ParseTime(parser, "displayTime")
	e.Type = parser.String("eventType")
	e.SubType = parser.String("eventSubType")

	if e.Type != nil {
		switch *e.Type {
		case EventTypeCarbs:
			e.Unit = StringOrDefault(parser, "unit", EventUnitCarbsGrams)
			e.Value = StringOrDefault(parser, "value", EventValueCarbsGramsMinimum)
		case EventTypeExercise:
			e.Unit = StringOrDefault(parser, "unit", EventUnitExerciseMinutes)
			e.Value = StringOrDefault(parser, "value", EventValueExerciseMinutesMinimum)
		case EventTypeInsulin:
			e.Unit = StringOrDefault(parser, "unit", EventUnitInsulinUnits)
			e.Value = StringOrDefault(parser, "value", EventValueInsulinUnitsMinimum)
		case EventTypeBG:
			e.Unit = StringOrDefault(parser, "unit", EventUnitMgdL)
			e.Value = StringOrDefault(parser, "value", EventValueMgdLMinimum)
		default:
			e.Unit = parser.String("unit")
			e.Value = parser.String("value")
		}
	}

	e.ID = parser.String("recordId")
	e.Status = parser.String("eventStatus")
	e.TransmitterGeneration = parser.String("transmitterGeneration")
	e.TransmitterID = parser.String("transmitterId")
	e.DisplayDevice = parser.String("displayDevice")
}

func (e *Event) Validate(validator structure.Validator) {
	validator = validator.WithMeta(e)

	validator.Time("systemTime", e.SystemTime.Raw()).NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", e.DisplayTime.Raw()).NotZero()
	validator.String("eventType", e.Type).Exists().OneOf(EventTypes()...)
	if e.Type != nil {
		switch *e.Type {
		case EventTypeCarbs:
			e.validateCarbs(validator)
		case EventTypeExercise:
			e.validateExercise(validator)
		case EventTypeHealth:
			e.validateHealth(validator)
		case EventTypeInsulin:
			e.validateInsulin(validator)
		case EventTypeNote, EventTypeNotes:
			e.validateNote(validator)
		case EventTypeBG:
			e.validateBG(validator)
		case EventTypeUnknown:
			e.validateUnknown(validator)
		}
	}
	validator.String("recordId", e.ID).Exists().NotEmpty()
	validator.String("eventStatus", e.Status).Exists().OneOf(EventStatuses()...)
	validator.String("transmitterId", e.TransmitterID).Exists().Using(TransmitterIDValidator)
	validator.String("transmitterGeneration", e.TransmitterGeneration).Exists().OneOf(DeviceTransmitterGenerations()...)
	validator.String("displayDevice", e.DisplayDevice).Exists().OneOf(DeviceDisplayDevices()...)
}

func (e *Event) validateCarbs(validator structure.Validator) {
	validator.String("eventSubType", e.SubType).NotExists()
	validator.String("unit", e.Unit).Exists().OneOf(EventUnitCarbsGrams)
	validator.String("value", e.Value).Exists().NotEmpty()
	if e.Value != nil {
		floatVal, err := strconv.ParseFloat(*e.Value, 64)
		if err != nil {
			floatVal = eventErrorValue
		}
		validator.Float64("value", &floatVal).Exists().InRange(EventValueCarbsGramsMinimum, EventValueCarbsGramsMaximum)
	}
}

func (e *Event) validateExercise(validator structure.Validator) {
	validator.String("eventSubType", e.SubType).OneOf(EventSubTypesExercise()...)
	validator.String("unit", e.Unit).Exists().OneOf(EventUnitExerciseMinutes)
	validator.String("value", e.Value).Exists().NotEmpty()
	if e.Value != nil {
		floatVal, err := strconv.ParseFloat(*e.Value, 64)
		if err != nil {
			floatVal = eventErrorValue
		}
		validator.Float64("value", &floatVal).Exists().InRange(EventValueExerciseMinutesMinimum, EventValueExerciseMinutesMaximum)
	}
}

func (e *Event) validateHealth(validator structure.Validator) {
	validator.String("eventSubType", e.SubType).OneOf(EventSubTypesHealth()...)
	validator.String("value", e.Value).Exists().NotEmpty()
}

func (e *Event) validateNote(validator structure.Validator) {
	validator.String("value", e.Value).Exists().NotEmpty()
}

func (e *Event) validateUnknown(validator structure.Validator) {
	validator.String("value", e.Value).Exists().NotEmpty()
}

func (e *Event) validateInsulin(validator structure.Validator) {
	validator.String("eventSubType", e.SubType).OneOf(EventSubTypesInsulin()...)
	validator.String("unit", e.Unit).Exists().OneOf(EventUnitInsulinUnits)
	validator.String("value", e.Value).Exists().NotEmpty()
	if e.Value != nil {
		floatVal, err := strconv.ParseFloat(*e.Value, 64)
		if err != nil {
			floatVal = eventErrorValue
		}
		validator.Float64("value", &floatVal).Exists().InRange(EventValueInsulinUnitsMinimum, EventValueInsulinUnitsMaximum)
	}
}

func (e *Event) validateBG(validator structure.Validator) {
	validator.String("unit", e.Unit).Exists().OneOf(EventUnitMgdL)
	validator.String("value", e.Value).Exists().NotEmpty()
	if e.Value != nil {
		floatVal, err := strconv.ParseFloat(*e.Value, 64)
		if err != nil {
			floatVal = eventErrorValue
		}
		validator.Float64("value", &floatVal).Exists().InRange(EGVValueMgdLMinimum, EGVValueMgdLMaximum)
	}
}
