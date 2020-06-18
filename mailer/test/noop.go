package test

import (
	"context"

	"github.com/tidepool-org/mailer/mailer"
)

type NoopMailer struct{}

func NewNoopMailer() mailer.Mailer {
	return &NoopMailer{}
}

func (n *NoopMailer) Send(ctx context.Context, email *mailer.Email) error {
	return nil
}
