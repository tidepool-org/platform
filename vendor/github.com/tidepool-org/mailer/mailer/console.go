package mailer

import (
	"context"
	"go.uber.org/zap"
)

const (
	ConsoleMailerBackendID = "console"
)

type ConsoleMailer struct {
	logger *zap.SugaredLogger
}

// Compiler time interface check
var _ Mailer = &ConsoleMailer{}

func NewConsoleMailer(logger *zap.SugaredLogger) *ConsoleMailer {
	return &ConsoleMailer{logger: logger}
}

func (c *ConsoleMailer) Send(ctx context.Context, email *Email) error {
	c.logger.Infow("Received new email message", "email", email)
	return nil
}
