package connectaccount

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/go-common/events"
	"github.com/tidepool-org/platform/conditionalnotifications"
	"github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
)

const (
	processorType            = "org.tidepool.processors.connect.account"
	quantity                 = 4
	frequency                = time.Minute
	processingTimeoutSeconds = 60
)

// NewGroupID returns a string suitable for [work.Work.GroupID] for batch deletions.
func NewGroupID(email, providerName string) string {
	return fmt.Sprintf("%s:%s:%s", processorType, email, providerName)
}

type Metadata struct {
	ClinicId          string    `json:"clinicId,omitempty"`
	Email             string    `json:"email,omitempty"`
	EmailTemplate     string    `json:"emailTemplate,omitempty"`
	PatientName       string    `json:"patientName,omitempty"`
	ProviderName      string    `json:"providerName,omitempty"`
	RestrictedTokenId string    `json:"restrictedTokenId,omitempty"`
	UserId            string    `json:"userId,omitempty"`
	WhenToSend        time.Time `json:"whenToSend,omitzero"`
}

type processor struct {
	dependencies conditionalnotifications.Dependencies
}

func NewWorkCreate(notBefore time.Time, metadata Metadata) *work.Create {
	return &work.Create{
		Type:                    processorType,
		SerialID:                pointer.FromString(metadata.UserId),
		GroupID:                 pointer.FromString(NewGroupID(metadata.Email, metadata.ProviderName)),
		ProcessingTimeout:       processingTimeoutSeconds,
		ProcessingAvailableTime: notBefore,
		Metadata:                fromConnectAccountData(metadata),
	}
}

func (d *Metadata) Parse(parser structure.ObjectParser) {
	d.ClinicId = pointer.ToString(parser.String("clinicId"))
	d.Email = pointer.ToString(parser.String("email"))
	d.EmailTemplate = pointer.ToString(parser.String("emailTemplate"))
	d.PatientName = pointer.ToString(parser.String("patientName"))
	d.ProviderName = pointer.ToString(parser.String("providerName"))
	d.RestrictedTokenId = pointer.ToString(parser.String("restrictedTokenId"))
	d.UserId = pointer.ToString(parser.String("userId"))
	d.WhenToSend = pointer.ToTime(parser.Time("whenToSend", time.RFC3339Nano))
}

func (d *Metadata) Validate(validator structure.Validator) {
	validator.String("clinicId", &d.ClinicId).NotEmpty()
	validator.String("email", &d.Email).NotEmpty()
	validator.String("emailTemplate", &d.EmailTemplate).NotEmpty()
	validator.String("patientName", &d.PatientName).NotEmpty()
	validator.String("providerName", &d.ProviderName).NotEmpty()
	validator.String("restrictedTokenId", &d.RestrictedTokenId).NotEmpty()
	validator.String("userId", &d.UserId).NotEmpty()
}

func NewProcessor(dependencies conditionalnotifications.Dependencies) *processor {
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
		return conditionalnotifications.NewFailingResult(err, wrk)
	}

	user, err := p.dependencies.Users.Get(ctx, data.UserId)
	if err != nil {
		return conditionalnotifications.NewFailingResult(err, wrk)
	}
	if user == nil || user.Username == nil {
		return conditionalnotifications.NewFailingResult(fmt.Errorf(`unable to find user for userId "%s"`, data.UserId), wrk)
	}
	filter := source.NewFilter()
	filter.ProviderName = pointer.FromStringArray([]string{data.ProviderName})
	filter.State = pointer.FromStringArray([]string{"connected"})
	connectedDataSources, err := p.dependencies.DataSources.List(ctx, data.UserId, filter, nil)
	if err != nil {
		return conditionalnotifications.NewFailingResult(err, wrk)
	}
	if len(connectedDataSources) > 0 {
		// User now has a connected dataSource so no email to send.
		return *work.NewProcessResultDelete()
	}

	emailVars := map[string]string{
		"RestrictedTokenId": data.RestrictedTokenId,
		"PatientName":       data.PatientName,
		"ProviderName":      data.ProviderName,
	}
	templateEvent := events.SendEmailTemplateEvent{
		Recipient: *user.Username,
		Template:  data.EmailTemplate,
		Variables: emailVars,
	}
	if err := p.dependencies.Mailer.SendEmailTemplate(ctx, templateEvent); err != nil {
		return conditionalnotifications.NewFailingResult(err, wrk)
	}
	return *work.NewProcessResultDelete()
}

func toConnectAccountData(wrk *work.Work) (*Metadata, error) {
	wrk.EnsureMetadata()
	var data Metadata
	if userId, ok := wrk.Metadata["userId"].(string); ok {
		data.UserId = userId
	} else {
		return nil, fmt.Errorf(`expected field "userId" to exist and be a string, received %T`, wrk.Metadata["userId"])
	}
	if providerName, ok := wrk.Metadata["providerName"].(string); ok {
		data.ProviderName = providerName
	} else {
		return nil, fmt.Errorf(`expected field "providerName" to exist and be a string, received %T`, wrk.Metadata["providerName"])
	}
	if patientName, ok := wrk.Metadata["patientName"].(string); ok {
		data.PatientName = patientName
	} else {
		return nil, fmt.Errorf(`expected field "patientName" to exist and be a string, received %T`, wrk.Metadata["patientName"])
	}
	if restrictedTokenId, ok := wrk.Metadata["restrictedTokenId"].(string); ok {
		data.RestrictedTokenId = restrictedTokenId
	} else {
		return nil, fmt.Errorf(`expected field "restrictedTokenId" to exist and be a string, received %T`, wrk.Metadata["restrictedTokenId"])
	}
	if emailTemplate, ok := wrk.Metadata["emailTemplate"].(string); ok {
		data.EmailTemplate = emailTemplate
	} else {
		return nil, fmt.Errorf(`expected field "emailTemplate" to exist and be a string, received %T`, wrk.Metadata["emailTemplate"])
	}
	return &data, nil
}

func fromConnectAccountData(data Metadata) map[string]any {
	return map[string]any{
		"userId":            data.UserId,
		"providerName":      data.ProviderName,
		"patientName":       data.PatientName,
		"restrictedTokenId": data.RestrictedTokenId,
		"emailTemplate":     data.EmailTemplate,
	}
}
