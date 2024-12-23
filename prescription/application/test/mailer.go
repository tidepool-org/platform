package test

import (
	"context"

	"github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/events"
)

type NoopMailer struct{}

var _ clients.MailerClient = &NoopMailer{}

func (n NoopMailer) SendEmailTemplate(ctx context.Context, event events.SendEmailTemplateEvent) error {
	return nil
}

func NewNoopMailer() clients.MailerClient {
	return &NoopMailer{}
}

//go:generate mockgen -source=mailer.go -destination=mock.go -package test MailerClient
type MailerClient interface {
	clients.MailerClient
}
