package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

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

type ConsentMailer struct {
	mailer     clients.MailerClient
	userClient user.Client
}

func NewConsentMailer(mailer clients.MailerClient, userClient user.Client) (*ConsentMailer, error) {
	return &ConsentMailer{
		mailer:     mailer,
		userClient: userClient,
	}, nil
}

func (c *ConsentMailer) SendConsentGrantedEmailNotification(ctx context.Context, cons consent.Consent, record consent.Record) error {
	usr, err := c.userClient.Get(ctx, record.UserID)
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

	return c.mailer.SendEmailTemplate(ctx, email)
}

func (c *ConsentMailer) SendConsentRevokedEmailNotification(ctx context.Context, cons consent.Consent, record consent.Record) error {
	usr, err := c.userClient.Get(ctx, record.UserID)
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

	return c.mailer.SendEmailTemplate(ctx, email)
}
