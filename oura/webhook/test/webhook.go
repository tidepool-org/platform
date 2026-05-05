package test

import (
	"github.com/tidepool-org/platform/oura"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomObjectID() string {
	return test.RandomStringFromCharset(test.CharsetAlphaNumeric)
}

func RandomEvent(options ...test.Option) *ouraWebhook.Event {
	return &ouraWebhook.Event{
		EventTime: pointer.From(test.RandomTime()),
		EventType: pointer.From(test.RandomStringFromArray(oura.EventTypes())),
		UserID:    pointer.From(ouraTest.RandomUserID()),
		ObjectID:  pointer.From(RandomObjectID()),
		DataType:  pointer.From(test.RandomStringFromArray(ouraWebhook.DataTypes())),
	}
}

func CloneEvent(datum *ouraWebhook.Event) *ouraWebhook.Event {
	if datum == nil {
		return nil
	}
	return &ouraWebhook.Event{
		EventTime: pointer.Clone(datum.EventTime),
		EventType: pointer.Clone(datum.EventType),
		UserID:    pointer.Clone(datum.UserID),
		ObjectID:  pointer.Clone(datum.ObjectID),
		DataType:  pointer.Clone(datum.DataType),
	}
}

func NewObjectFromEvent(datum *ouraWebhook.Event, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.EventTime != nil {
		object["event_time"] = test.NewObjectFromTime(*datum.EventTime, format)
	}
	if datum.EventType != nil {
		object["event_type"] = test.NewObjectFromString(*datum.EventType, format)
	}
	if datum.UserID != nil {
		object["user_id"] = test.NewObjectFromString(*datum.UserID, format)
	}
	if datum.ObjectID != nil {
		object["object_id"] = test.NewObjectFromString(*datum.ObjectID, format)
	}
	if datum.DataType != nil {
		object["data_type"] = test.NewObjectFromString(*datum.DataType, format)
	}
	return object
}

func RandomEventMetadata(options ...test.Option) *ouraWebhook.EventMetadata {
	return &ouraWebhook.EventMetadata{
		Event: test.RandomOptionalPointerWithOptions(RandomEvent, options...),
	}
}

func CloneEventMetadata(datum *ouraWebhook.EventMetadata) *ouraWebhook.EventMetadata {
	if datum == nil {
		return nil
	}
	return &ouraWebhook.EventMetadata{
		Event: CloneEvent(datum.Event),
	}
}

func NewObjectFromEventMetadata(datum *ouraWebhook.EventMetadata, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.Event != nil {
		object["event"] = NewObjectFromEvent(datum.Event, format)
	}
	return object
}
