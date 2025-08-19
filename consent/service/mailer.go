package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	"github.com/tidepool-org/platform/mailer"

	"github.com/tidepool-org/platform/log"

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/go-common/clients"
	"github.com/tidepool-org/go-common/events"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/user"
)

const (
	defaultInformedConsentGrantedEmailTemplate = "informed_consent_granted"
	defaultInformedConsentRevokedEmailTemplate = "informed_consent_revoked"
)

var (
	customInformedConsentTemplates = map[string]string{
		consent.TypeBigDataDonationProject: "informed_consent_granted_big_data_donation_project",
	}
)

type ConsentMailerConfig struct {
	Disabled bool `envconfig:"TIDEPOOL_CONSENT_MAILER_DISABLED"`
}

type ConsentMailer interface {
	SendConsentGrantedEmailNotification(ctx context.Context, cons consent.Consent, record consent.Record) error
	SendConsentRevokedEmailNotification(ctx context.Context, cons consent.Consent, record consent.Record) error
}

type defaultConsentMailer struct {
	logger     log.Logger
	mailer     clients.MailerClient
	userClient user.Client
}

func NewConsentMailer(userClient user.Client, logger log.Logger) (ConsentMailer, error) {
	config := &ConsentMailerConfig{}
	err := envconfig.Process("", config)
	if err != nil {
		return nil, err
	}

	if config.Disabled {
		return &disabledConsentMailer{
			logger: logger,
		}, nil
	}

	mlr, err := mailer.Client()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create mailer client")
	}

	return &defaultConsentMailer{
		logger:     logger,
		mailer:     mlr,
		userClient: userClient,
	}, nil
}

func (d *defaultConsentMailer) SendConsentGrantedEmailNotification(ctx context.Context, cons consent.Consent, record consent.Record) error {
	usr, err := d.userClient.Get(ctx, record.UserID)
	if err != nil || usr == nil {
		return errors.Wrap(err, "could not get user")
	}

	if usr.Username == nil || *usr.Username == "" {
		return nil
	}

	renderer, err := NewMarkdownConsentRenderer(cons, record)
	if err != nil {
		return errors.Wrap(err, "could not create renderer")
	}

	var buffer bytes.Buffer
	err = renderer.RenderPDF(&buffer)
	if err != nil {
		return errors.Wrap(err, "could not render consent PDF")
	}
	encoded := base64.StdEncoding.EncodeToString(buffer.Bytes())

	template := defaultInformedConsentGrantedEmailTemplate
	if customTemplate, exists := customInformedConsentTemplates[cons.Type]; exists {
		template = customTemplate
	}
	email := events.SendEmailTemplateEvent{
		Template:  template,
		Recipient: *usr.Username,
		Variables: map[string]string{
			"Name": record.OwnerName,
		},
		Attachments: []events.EmailAttachment{{
			ContentType: "application/pdf",
			Data:        encoded,
			Filename:    fmt.Sprintf("%s.v%d.pdf", record.Type, record.Version),
		}},
	}

	return d.mailer.SendEmailTemplate(ctx, email)
}

func (d *defaultConsentMailer) SendConsentRevokedEmailNotification(ctx context.Context, cons consent.Consent, record consent.Record) error {
	usr, err := d.userClient.Get(ctx, record.UserID)
	if err != nil || usr == nil {
		return errors.Wrap(err, "could not get user")
	}

	if usr.Username == nil || *usr.Username == "" {
		return nil
	}

	email := events.SendEmailTemplateEvent{
		Template:  defaultInformedConsentRevokedEmailTemplate,
		Recipient: *usr.Username,
		Variables: map[string]string{
			"Name": record.OwnerName,
			"Type": consent.PrettifyType(record.Type),
		},
	}

	return d.mailer.SendEmailTemplate(ctx, email)
}

type disabledConsentMailer struct {
	logger log.Logger
}

func (d *disabledConsentMailer) SendConsentGrantedEmailNotification(_ context.Context, _ consent.Consent, record consent.Record) error {
	d.logger.WithFields(log.Fields{"userId": record.UserID}).WithError(errors.New("consent mailer is disabled")).Info("SendConsentGrantedNotification")
	return nil
}

func (d *disabledConsentMailer) SendConsentRevokedEmailNotification(_ context.Context, _ consent.Consent, record consent.Record) error {
	d.logger.WithFields(log.Fields{"userId": record.UserID}).WithError(errors.New("consent mailer is disabled")).Info("SendConsentRevokedEmailNotification")
	return nil
}
