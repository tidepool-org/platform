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

//go:generate mockgen --build_flags=--mod=mod -source=./mailer.go -destination=./mock.go -package test MockMailer

type MockMailer interface {
	clients.MailerClient
}
