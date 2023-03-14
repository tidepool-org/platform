package dexcom

import (
	"strconv"

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
	EventTypeNotes    = "notes"

	EventUnitUnknown = "unknown"
	EventUnitMgdL    = "mg/dL"

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
)

func EventTypes() []string {
	return []string{
		EventTypeBG,
		EventTypeCarbs,
		EventTypeExercise,
		EventTypeHealth,
		EventTypeInsulin,
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
	Events *Events `json:"records,omitempty"`
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
	SystemTime            *Time    `json:"systemTime,omitempty"`
	DisplayTime           *Time    `json:"displayTime,omitempty"`
	Type                  *string  `json:"eventType,omitempty"`
	SubType               *string  `json:"eventSubType,omitempty"`
	Unit                  *string  `json:"unit,omitempty"`
	Value                 *float64 `json:"value,omitempty"`
	ID                    *string  `json:"recordId,omitempty"`
	Status                *string  `json:"eventStatus,omitempty"`
	TransmitterId         *string  `json:"transmitterId,omitempty"`
	TransmitterGeneration *string  `json:"transmitterGeneration,omitempty"`
	DisplayDevice         *string  `json:"displayDevice,omitempty"`
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
	e.SystemTime = TimeFromRaw(parser.Time("systemTime", TimeFormat))
	e.DisplayTime = TimeFromRaw(parser.Time("displayTime", TimeFormat))
	e.Type = parser.String("eventType")
	e.SubType = parser.String("eventSubType")
	e.Unit = parser.String("unit")
	e.Value = parser.Float64("value")
	e.ID = parser.String("recordId")
	e.Status = parser.String("eventStatus")
	e.TransmitterGeneration = parser.String("transmitterGeneration")
	e.TransmitterId = parser.String("transmitterId")
	e.DisplayDevice = parser.String("displayDevice")
}

func (e *Event) Validate(validator structure.Validator) {
	validator = validator.WithMeta(e)
	validator.Time("systemTime", e.SystemTime.Raw()).Exists().NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", e.DisplayTime.Raw()).Exists().NotZero()
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
		}
	}
	validator.String("recordId", e.ID).Exists().NotEmpty()
	validator.String("eventStatus", e.Status).Exists().OneOf(EventStatuses()...)

	validator.String("transmitterId", e.TransmitterId).Exists().NotEmpty()
	validator.String("transmitterGeneration", e.TransmitterGeneration).Exists().OneOf(DeviceTransmitterGenerations()...)
	validator.String("displayDevice", e.DisplayDevice).Exists().OneOf(DeviceDisplayDevices()...)
}

func (e *Event) validateCarbs(validator structure.Validator) {
	validator.String("eventSubType", e.SubType).NotExists()
	if e.Unit != nil || e.Value != nil {
		validator.String("unit", e.Unit).Exists().OneOf(EventUnitCarbsGrams)
		validator.Float64("value", e.Value).Exists().InRange(EventValueCarbsGramsMinimum, EventValueCarbsGramsMaximum)
	}
}

func (e *Event) validateExercise(validator structure.Validator) {
	// HACK: Dexcom - value of -1 is invalid; ignore unit and value instead (per Dexcom)
	if e.Value != nil && *e.Value == -1.0 {
		e.Unit = nil
		e.Value = nil
	}

	validator.String("eventSubType", e.SubType).OneOf(EventSubTypesExercise()...)
	if e.Unit != nil || e.Value != nil {
		validator.String("unit", e.Unit).Exists().OneOf(EventUnitExerciseMinutes)
		validator.Float64("value", e.Value).Exists().InRange(EventValueExerciseMinutesMinimum, EventValueExerciseMinutesMaximum)
	}
}

func (e *Event) validateHealth(validator structure.Validator) {
	validator.String("eventSubType", e.SubType).OneOf(EventSubTypesHealth()...)
	validator.String("unit", e.Unit).NotExists()
	validator.Float64("value", e.Value).EqualTo(0)
}

func (e *Event) validateInsulin(validator structure.Validator) {
	validator.String("eventSubType", e.SubType).OneOf(EventSubTypesInsulin()...)
	if e.Unit != nil || e.Value != nil {
		validator.String("unit", e.Unit).Exists().OneOf(EventUnitInsulinUnits)
		validator.Float64("value", e.Value).Exists().InRange(EventValueInsulinUnitsMinimum, EventValueInsulinUnitsMaximum)
	}
}
