package mailer

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	DefaultCharset = "UTF-8"
)

type Email struct {
	Recipients []string `json:"recipients" validate:"min=1,email"`
	Cc         []string `json:"cc" validate:"email"`
	Subject    string   `json:"subject" validate:"required"`
	Body       string   `json:"body" validate:"required"`
}

type Mailer interface {
	Send(ctx context.Context, email *Email) error
}

func New(id string, logger *zap.SugaredLogger, validate *validator.Validate) (Mailer, error) {
	switch id {
	case SESMailerBackendID:
		logger.Info("Creating new ses mailer backend")
		backendConfig := &SESMailerConfig{}
		if err := envconfig.Process("", backendConfig); err != nil {
			return nil, err
		}
		if err := validate.Struct(backendConfig); err != nil {
			return nil, err
		}

		params := &SESMailerParams{
			Cfg: backendConfig,
			Logger: logger,
		}
		return NewSESMailer(params)
	case ConsoleMailerBackendID:
		logger.Info("Creating new console mailer backend")
		return NewConsoleMailer(logger), nil
	default:
		return nil, errors.New(fmt.Sprintf("unknown mailer backend %s", id))
	}
}