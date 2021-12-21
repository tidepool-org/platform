package clients

import (
	"context"
	"github.com/tidepool-org/go-common/events"
)

const MailerTopic = "emails"

type MailerClient interface {
	SendEmailTemplate(context.Context, events.SendEmailTemplateEvent) error
}

type mailerClient struct {
	producer *events.KafkaCloudEventsProducer
}

func NewMailerClient(config *events.CloudEventsConfig) (MailerClient, error) {
	producer, err := events.NewKafkaCloudEventsProducer(config)
	if err != nil {
		return nil, err
	}

	return &mailerClient{producer}, nil
}

func (m *mailerClient) SendEmailTemplate(ctx context.Context, event events.SendEmailTemplateEvent) error {
	return m.producer.Send(ctx, event)
}
