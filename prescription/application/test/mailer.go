package test

import (
	"context"

	"github.com/tidepool-org/go-common/events"

	"github.com/tidepool-org/platform/mailer"
)

//go:generate mockgen -source=mailer.go -destination=mailer_mocks.go -package=test -typed

type NoopMailer struct{}

var _ mailer.Client = &NoopMailer{}

func (n NoopMailer) SendEmailTemplate(ctx context.Context, event events.SendEmailTemplateEvent) error {
	return nil
}

func NewNoopMailer() mailer.Client {
	return &NoopMailer{}
}

type MailerClient interface {
	mailer.Client
}
