package events

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/tidepool-org/go-common/clients/shoreline"
	"log"
)

const (
	DeleteUserEventType = "users:delete"
	UpdateUserEventType = "users:update"
	CreateUserEventType = "users:create"
)

type Event interface {
	GetEventType() string
	GetEventKey() string
}

var _ Event = DeleteUserEvent{}

type DeleteUserEvent struct {
	shoreline.UserData `json:",inline"`
}

func (d DeleteUserEvent) GetEventType() string {
	return DeleteUserEventType
}

func (d DeleteUserEvent) GetEventKey() string {
	return d.UserID
}

var _ Event = CreateUserEvent{}

type CreateUserEvent struct {
	shoreline.UserData `json:",inline"`
}

func (c CreateUserEvent) GetEventType() string {
	return CreateUserEventType
}

func (c CreateUserEvent) GetEventKey() string {
	return c.UserID
}

var _ Event = UpdateUserEvent{}

type UpdateUserEvent struct {
	Original shoreline.UserData `json:"original"`
	Updated  shoreline.UserData `json:"updated"`
}

func (u UpdateUserEvent) GetEventType() string {
	return UpdateUserEventType
}

func (u UpdateUserEvent) GetEventKey() string {
	return u.Original.UserID
}

type UserEventsHandler interface {
	HandleUpdateUserEvent(payload UpdateUserEvent) error
	HandleCreateUserEvent(payload CreateUserEvent) error
	HandleDeleteUserEvent(payload DeleteUserEvent) error
}

func NewUserEventsHandler(delegate UserEventsHandler) EventHandler {
	return &delegatingUserEventsHandler{delegate: delegate}
}

var _ EventHandler = &delegatingUserEventsHandler{}
type delegatingUserEventsHandler struct {
	delegate UserEventsHandler
}

func (d *delegatingUserEventsHandler) CanHandle(ce cloudevents.Event) bool {
	switch ce.Type() {
	case CreateUserEventType, UpdateUserEventType, DeleteUserEventType:
		return true
	default:
		return false
	}
}

func (d *delegatingUserEventsHandler) Handle(ce cloudevents.Event) error {
	switch ce.Type() {
	case CreateUserEventType:
		payload := CreateUserEvent{}
		if err := ce.DataAs(&payload); err != nil {
			return err
		}
		return d.delegate.HandleCreateUserEvent(payload)
	case UpdateUserEventType:
		payload := UpdateUserEvent{}
		if err := ce.DataAs(&payload); err != nil {
			return err
		}
		return d.delegate.HandleUpdateUserEvent(payload)
	case DeleteUserEventType:
		payload := DeleteUserEvent{}
		if err := ce.DataAs(&payload); err != nil {
			return err
		}
		return d.delegate.HandleDeleteUserEvent(payload)
	}
	return nil
}

type DebugEventHandler struct {}
var _ EventHandler = &DebugEventHandler{}

func (d *DebugEventHandler) CanHandle(ce cloudevents.Event) bool {
	return true
}

func (d *DebugEventHandler) Handle(ce cloudevents.Event) error {
	log.Printf("Received event %v\n", ce)
	return nil
}

type NoopUserEventsHandler struct{}
var _ UserEventsHandler = &NoopUserEventsHandler{}

func (d *NoopUserEventsHandler) HandleUpdateUserEvent(payload UpdateUserEvent) error {
	return nil
}
func (d *NoopUserEventsHandler) HandleCreateUserEvent(payload CreateUserEvent) error {
	return nil
}
func (d *NoopUserEventsHandler) HandleDeleteUserEvent(payload DeleteUserEvent) error {
	return nil
}
