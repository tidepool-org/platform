package events

import cloudevents "github.com/cloudevents/sdk-go/v2"

const (
	SendEmailTemplateEventType = "email_template:send"
)

type SendEmailTemplateEvent struct {
	Recipient string            `json:"recipient"`
	Template  string            `json:"template"`
	Variables map[string]string `json:"variables"`
}

func (s SendEmailTemplateEvent) GetEventType() string {
	return SendEmailTemplateEventType
}

func (s SendEmailTemplateEvent) GetEventKey() string {
	return s.Recipient
}

var _ Event = &SendEmailTemplateEvent{}

type EmailEventHandler interface {
	HandleSendEmailTemplate(payload SendEmailTemplateEvent) error
}

type DelegatingEmailEventHandler struct {
	delegate EmailEventHandler
}

func NewDelegatingEmailEventHandler(delegate EmailEventHandler) *DelegatingEmailEventHandler {
	return &DelegatingEmailEventHandler{delegate: delegate}
}

func (d DelegatingEmailEventHandler) CanHandle(ce cloudevents.Event) bool {
	return ce.Type() == SendEmailTemplateEventType
}

func (d DelegatingEmailEventHandler) Handle(ce cloudevents.Event) error {
	if ce.Type() != SendEmailTemplateEventType {
		// ignore invalid events
		return nil
	}

	payload := SendEmailTemplateEvent{}
	if err := ce.DataAs(&payload); err != nil {
		return err
	}
	return d.delegate.HandleSendEmailTemplate(payload)
}

var _ EventHandler = &DelegatingEmailEventHandler{}
