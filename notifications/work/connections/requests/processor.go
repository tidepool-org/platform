package requests

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/go-common/events"
	"github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/notifications"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
)

const (
	processorType            = "org.tidepool.processors.connections.requests"
	quantity                 = 2
	frequency                = time.Minute
	processingTimeoutSeconds = 60
)

// NewGroupID returns a string suitable for [work.Work.GroupID] for batch deletions.
func NewGroupID(userID, providerName string) string {
	return fmt.Sprintf("%s:%s:%s", processorType, userID, providerName)
}

type Metadata struct {
	ClinicID          string    `json:"clinicId,omitempty"`
	Email             string    `json:"email,omitempty"`
	EmailTemplate     string    `json:"emailTemplate,omitempty"`
	PatientName       string    `json:"patientName,omitempty"`
	ProviderName      string    `json:"providerName,omitempty"`
	RestrictedTokenID string    `json:"restrictedTokenId,omitempty"`
	UserID            string    `json:"userId,omitempty"`
	WhenToSend        time.Time `json:"whenToSend,omitzero"`
}

type processor struct {
	dependencies notifications.Dependencies
}

func AddWorkItem(ctx context.Context, client work.Client, metadata Metadata) error {
	whenToSend := metadata.WhenToSend
	if whenToSend.IsZero() {
		whenToSend = time.Now().Add(time.Hour * 24 * 7)
	}
	create := newWorkCreate(whenToSend, metadata)
	if groupID := pointer.DefaultString(create.GroupID, ""); groupID != "" {
		if _, err := client.DeleteAllByGroupID(ctx, groupID); err != nil {
			return errors.Wrapf(err, `unable to delete existing groups by id "%s"`, groupID)
		}
	}
	if _, err := client.Create(ctx, create); err != nil {
		return err
	}
	return nil
}

func newWorkCreate(notBefore time.Time, metadata Metadata) *work.Create {
	return &work.Create{
		Type:                    processorType,
		SerialID:                pointer.FromString(metadata.UserID),
		GroupID:                 pointer.FromString(NewGroupID(metadata.UserID, metadata.ProviderName)),
		ProcessingTimeout:       processingTimeoutSeconds,
		ProcessingAvailableTime: notBefore,
		Metadata:                fromConnectAccountData(metadata),
	}
}

func (d *Metadata) Parse(parser structure.ObjectParser) {
	d.ClinicID = pointer.ToString(parser.String("clinicId"))
	d.Email = pointer.ToString(parser.String("email"))
	d.EmailTemplate = pointer.ToString(parser.String("emailTemplate"))
	d.PatientName = pointer.ToString(parser.String("patientName"))
	d.ProviderName = pointer.ToString(parser.String("providerName"))
	d.RestrictedTokenID = pointer.ToString(parser.String("restrictedTokenId"))
	d.UserID = pointer.ToString(parser.String("userId"))
	d.WhenToSend = pointer.ToTime(parser.Time("whenToSend", time.RFC3339Nano))
}

func (d *Metadata) Validate(validator structure.Validator) {
	validator.String("clinicId", &d.ClinicID).NotEmpty()
	validator.String("email", &d.Email).NotEmpty()
	validator.String("emailTemplate", &d.EmailTemplate).NotEmpty()
	validator.String("patientName", &d.PatientName).NotEmpty()
	validator.String("providerName", &d.ProviderName).NotEmpty()
	validator.String("restrictedTokenId", &d.RestrictedTokenID).NotEmpty()
	validator.String("userId", &d.UserID).NotEmpty()
}

func NewProcessor(dependencies notifications.Dependencies) *processor {
	return &processor{
		dependencies: dependencies,
	}
}

func (p *processor) Type() string {
	return processorType
}

func (p *processor) Quantity() int {
	return quantity
}

func (p *processor) Frequency() time.Duration {
	return frequency
}

func (p *processor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) work.ProcessResult {
	data, err := toConnectAccountData(wrk)
	if err != nil {
		return notifications.NewFailingResult(err, wrk)
	}

	user, err := p.dependencies.Users.Get(ctx, data.UserID)
	if err != nil {
		return notifications.NewFailingResult(err, wrk)
	}
	if user == nil || user.Username == nil {
		return notifications.NewFailingResult(errors.Newf(`unable to find user for userId "%s"`, data.UserID), wrk)
	}
	filter := source.NewFilter()
	filter.ProviderName = pointer.FromStringArray([]string{data.ProviderName})
	filter.State = pointer.FromStringArray([]string{source.StateConnected})
	connectedDataSources, err := p.dependencies.DataSources.List(ctx, data.UserID, filter, nil)
	if err != nil {
		return notifications.NewFailingResult(err, wrk)
	}
	if len(connectedDataSources) > 0 {
		// User now has a connected dataSource so no email to send.
		return *work.NewProcessResultDelete()
	}

	var clinicName string
	clinic, err := p.dependencies.Clinics.GetClinic(ctx, data.ClinicID)
	if err != nil {
		return notifications.NewFailingResult(errors.Wrapf(err, `error getting clinic`), wrk)
	}
	if clinic != nil {
		clinicName = clinic.Name
	}
	emailVars := map[string]string{
		"ClinicName":        clinicName,
		"PatientName":       data.PatientName,
		"ProviderName":      data.ProviderName,
		"RestrictedTokenId": data.RestrictedTokenID,
	}
	templateEvent := events.SendEmailTemplateEvent{
		Recipient: *user.Username,
		Template:  data.EmailTemplate,
		Variables: emailVars,
	}
	if err := p.dependencies.Mailer.SendEmailTemplate(ctx, templateEvent); err != nil {
		return notifications.NewFailingResult(err, wrk)
	}
	return *work.NewProcessResultDelete()
}

func toConnectAccountData(wrk *work.Work) (*Metadata, error) {
	wrk.EnsureMetadata()
	var data Metadata
	if userID, ok := wrk.Metadata["userId"].(string); ok {
		data.UserID = userID
	} else {
		return nil, errors.Newf(`expected field "userId" to exist and be a string, received %T`, wrk.Metadata["userId"])
	}
	if providerName, ok := wrk.Metadata["providerName"].(string); ok {
		data.ProviderName = providerName
	} else {
		return nil, errors.Newf(`expected field "providerName" to exist and be a string, received %T`, wrk.Metadata["providerName"])
	}
	if patientName, ok := wrk.Metadata["patientName"].(string); ok {
		data.PatientName = patientName
	} else {
		return nil, errors.Newf(`expected field "patientName" to exist and be a string, received %T`, wrk.Metadata["patientName"])
	}
	if restrictedTokenID, ok := wrk.Metadata["restrictedTokenId"].(string); ok {
		data.RestrictedTokenID = restrictedTokenID
	} else {
		return nil, errors.Newf(`expected field "restrictedTokenId" to exist and be a string, received %T`, wrk.Metadata["restrictedTokenId"])
	}
	if emailTemplate, ok := wrk.Metadata["emailTemplate"].(string); ok {
		data.EmailTemplate = emailTemplate
	} else {
		return nil, errors.Newf(`expected field "emailTemplate" to exist and be a string, received %T`, wrk.Metadata["emailTemplate"])
	}
	return &data, nil
}

func fromConnectAccountData(data Metadata) map[string]any {
	return map[string]any{
		"userId":            data.UserID,
		"providerName":      data.ProviderName,
		"patientName":       data.PatientName,
		"restrictedTokenId": data.RestrictedTokenID,
		"emailTemplate":     data.EmailTemplate,
	}
}
