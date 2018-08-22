package dexcom

import (
	"strconv"
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type EventsResponse struct {
	Events *Events `json:"events,omitempty"`
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
	e.Events = ParseEvents(parser.WithReferenceArrayParser("events"))
}

func (e *EventsResponse) Validate(validator structure.Validator) {
	if eventsValidator := validator.WithReference("events"); e.Events != nil {
		eventsValidator.Validate(e.Events)
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
			eventValidator.Validate(event)
		} else {
			eventValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type Event struct {
	SystemTime   *time.Time `json:"systemTime,omitempty"`
	DisplayTime  *time.Time `json:"displayTime,omitempty"`
	EventType    *string    `json:"eventType,omitempty"`
	EventSubType *string    `json:"eventSubType,omitempty"`
	Unit         *string    `json:"unit,omitempty"`
	Value        *float64   `json:"value,omitempty"`
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
	e.SystemTime = parser.Time("systemTime", DateTimeFormat)
	e.DisplayTime = parser.Time("displayTime", DateTimeFormat)
	e.EventType = parser.String("eventType")
	e.EventSubType = parser.String("eventSubType")
	e.Unit = parser.String("unit")
	e.Value = parser.Float64("value")
}

func (e *Event) Validate(validator structure.Validator) {
	validator = validator.WithMeta(e)
	validator.Time("systemTime", e.SystemTime).Exists().NotZero().BeforeNow(NowThreshold)
	validator.Time("displayTime", e.DisplayTime).Exists().NotZero()
	validator.String("eventType", e.EventType).Exists().OneOf(EventCarbs, EventExercise, EventHealth, EventInsulin)
	if e.EventType != nil {
		switch *e.EventType {
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
}

func (e *Event) validateCarbs(validator structure.Validator) {
	validator.String("eventSubType", e.EventSubType).NotExists()
	if e.Unit != nil || e.Value != nil {
		validator.String("unit", e.Unit).Exists().EqualTo(UnitGrams)
		validator.Float64("value", e.Value).Exists().InRange(0, 250)
	}
}

func (e *Event) validateExercise(validator structure.Validator) {
	// HACK: Dexcom - value of -1 is invalid; ignore unit and value instead (per Dexcom)
	if e.Value != nil && *e.Value == -1.0 {
		e.Unit = nil
		e.Value = nil
	}

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
