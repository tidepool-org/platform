package dexcom

import (
	"strconv"
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type EventsResponse struct {
	Events []*Event `json:"events,omitempty"`
}

func NewEventsResponse() *EventsResponse {
	return &EventsResponse{}
}

func (e *EventsResponse) Parse(parser structure.ObjectParser) {
	if eventsParser := parser.WithReferenceArrayParser("events"); eventsParser.Exists() {
		for _, reference := range eventsParser.References() {
			if eventParser := eventsParser.WithReferenceObjectParser(reference); eventParser.Exists() {
				event := NewEvent()
				event.Parse(eventParser)
				eventParser.NotParsed()
				e.Events = append(e.Events, event)
			}
		}
		eventsParser.NotParsed()
	}
}

func (e *EventsResponse) Validate(validator structure.Validator) {
	validator = validator.WithReference("events")
	for index, event := range e.Events {
		if eventValidator := validator.WithReference(strconv.Itoa(index)); event != nil {
			event.Validate(eventValidator)
		} else {
			eventValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type Event struct {
	SystemTime   time.Time `json:"systemTime,omitempty"`
	DisplayTime  time.Time `json:"displayTime,omitempty"`
	EventType    string    `json:"eventType,omitempty"`
	EventSubType *string   `json:"eventSubType,omitempty"`
	Unit         *string   `json:"unit,omitempty"`
	Value        *float64  `json:"value,omitempty"`
}

func NewEvent() *Event {
	return &Event{}
}

func (e *Event) Parse(parser structure.ObjectParser) {
	if ptr := parser.Time("systemTime", DateTimeFormat); ptr != nil {
		e.SystemTime = *ptr
	}
	if ptr := parser.Time("displayTime", DateTimeFormat); ptr != nil {
		e.DisplayTime = *ptr
	}
	if ptr := parser.String("eventType"); ptr != nil {
		e.EventType = *ptr
	}
	e.EventSubType = parser.String("eventSubType")
	e.Unit = parser.String("unit")
	e.Value = parser.Float64("value")
}

func (e *Event) Validate(validator structure.Validator) {
	validator.Time("systemTime", &e.SystemTime).BeforeNow(NowThreshold)
	validator.Time("displayTime", &e.DisplayTime).NotZero()
	validator.String("eventType", &e.EventType).OneOf(EventCarbs, EventExercise, EventHealth, EventInsulin)

	switch e.EventType {
	case EventCarbs:
		e.validateCarbs(validator)
	case EventExercise:
		e.validateExercise(validator)
	case EventHealth:
		e.validateHealth(validator)
	case EventInsulin:
		e.validateInsulin(validator)
	}
}

func (e *Event) validateCarbs(validator structure.Validator) {
	validator.String("eventSubType", e.EventSubType).NotExists()
	if e.Unit != nil || e.Value != nil {
		validator.String("unit", e.Unit).Exists().EqualTo(UnitGrams)
		validator.Float64("value", e.Value).Exists().InRange(0, 250)
	}
}

func (e *Event) validateExercise(validator structure.Validator) {
	validator.String("eventSubType", e.EventSubType).OneOf(ExerciseLight, ExerciseMedium, ExerciseHeavy)
	if e.Unit != nil || e.Value != nil {
		validator.String("unit", e.Unit).Exists().EqualTo(UnitMinutes)
		validator.Float64("value", e.Value).Exists().InRange(0, 360)
	}
}

func (e *Event) validateHealth(validator structure.Validator) {
	validator.String("eventSubType", e.EventSubType).OneOf(HealthIllness, HealthStress, HealthHighSymptoms, HealthLowSymptoms, HealthCycle, HealthAlcohol)
	validator.String("unit", e.Unit).NotExists()
	validator.Float64("value", e.Value).NotExists()
}

func (e *Event) validateInsulin(validator structure.Validator) {
	validator.String("eventSubType", e.EventSubType).NotExists()
	if e.Unit != nil || e.Value != nil {
		validator.String("unit", e.Unit).Exists().EqualTo(UnitUnits)
		validator.Float64("value", e.Value).Exists().InRange(0, 250)
	}
}
