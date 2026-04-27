package webhook

import (
	"slices"
	"strings"
	"time"

	"github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

const EventPath = "/event"

// All except for heartrate (not available via webhook)
func DataTypes() []string {
	return slices.DeleteFunc(oura.DataTypes(), func(dataType string) bool {
		return dataType == oura.DataTypeHeartRate
	})
}

type Event struct {
	EventTime *time.Time `json:"event_time,omitempty" bson:"event_time,omitempty"`
	EventType *string    `json:"event_type,omitempty" bson:"event_type,omitempty"`
	UserID    *string    `json:"user_id,omitempty" bson:"user_id,omitempty"`
	ObjectID  *string    `json:"object_id,omitempty" bson:"object_id,omitempty"`
	DataType  *string    `json:"data_type,omitempty" bson:"data_type,omitempty"`
}

func ParseEvent(parser structure.ObjectParser) *Event {
	if !parser.Exists() {
		return nil
	}
	datum := &Event{}
	datum.Parse(parser)
	return datum
}

func (e *Event) Parse(parser structure.ObjectParser) {
	e.EventTime = parser.Time("event_time", time.RFC3339Nano)
	e.EventType = parser.String("event_type")
	e.UserID = parser.String("user_id")
	e.ObjectID = parser.String("object_id")
	e.DataType = parser.String("data_type")
}

func (e *Event) Validate(validator structure.Validator) {
	validator.Time("event_time", e.EventTime).Exists().NotZero()
	validator.String("event_type", e.EventType).Exists().OneOf(oura.EventTypes()...)
	validator.String("user_id", e.UserID).Exists().NotEmpty()
	validator.String("object_id", e.ObjectID).Exists().NotEmpty()
	validator.String("data_type", e.DataType).Exists().OneOf(DataTypes()...)
}

func (e *Event) String() string {
	var parts []string
	if e.EventTime != nil {
		parts = append(parts, e.EventTime.Format(time.RFC3339Nano))
	} else {
		parts = append(parts, "")
	}
	parts = append(parts,
		pointer.Default(e.EventType, ""),
		pointer.Default(e.UserID, ""),
		pointer.Default(e.ObjectID, ""),
		pointer.Default(e.DataType, ""),
	)
	return strings.Join(parts, ":")
}

const MetadataKeyEvent = "event"

type EventMetadata struct {
	Event *Event `json:"event,omitempty" bson:"event,omitempty"`
}

func (e *EventMetadata) Parse(parser structure.ObjectParser) {
	e.Event = ParseEvent(parser.WithReferenceObjectParser(MetadataKeyEvent))
}

func (e *EventMetadata) Validate(validator structure.Validator) {
	if e.Event != nil {
		e.Event.Validate(validator.WithReference(MetadataKeyEvent))
	}
}
