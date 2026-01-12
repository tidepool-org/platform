package issues

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/go-common/events"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"

	"github.com/tidepool-org/platform/notifications"
	"github.com/tidepool-org/platform/structure"
)

const (
	processorType            = "org.tidepool.processors.connections.issues"
	quantity                 = 2
	frequency                = time.Minute
	processingTimeoutSeconds = 60
)

// NewGroupID returns a string suitable for [work.Work.GroupID] for batch deletions.
func NewGroupID(dataSourceId string) string {
	return fmt.Sprintf("%s:%s", processorType, dataSourceId)
}

type processor struct {
	dependencies conditionalnotifications.Dependencies
}

type Metadata struct {
	DataSourceState   string `json:"dataSourceState,omitempty"`
	DataSourceId      string `json:"dataSourceId,omitempty"`
	EmailTemplate     string `json:"emailTemplate,omitempty"`
	PatientName       string `json:"patientName,omitempty"`
	ProviderName      string `json:"providerName,omitempty"`
	RestrictedTokenId string `json:"restrictedTokenId,omitempty"`
	UserId            string `json:"userId,omitempty"`
}

func (d *Metadata) Parse(parser structure.ObjectParser) {
	d.DataSourceState = pointer.ToString(parser.String("dataSourceState"))
	d.DataSourceId = pointer.ToString(parser.String("dataSourceId"))
	d.EmailTemplate = pointer.ToString(parser.String("emailTemplate"))
	d.PatientName = pointer.ToString(parser.String("patientName"))
	d.ProviderName = pointer.ToString(parser.String("providerName"))
	d.RestrictedTokenId = pointer.ToString(parser.String("restrictedTokenId"))
	d.UserId = pointer.ToString(parser.String("userId"))
}

func (d *Metadata) Validate(validator structure.Validator) {
	validator.String("dataSourceState", &d.DataSourceState).NotEmpty()
	validator.String("dataSourceId", &d.DataSourceId).NotEmpty()
	validator.String("emailTemplate", &d.EmailTemplate).NotEmpty()
	validator.String("patientName", &d.PatientName).NotEmpty()
	validator.String("providerName", &d.ProviderName).NotEmpty()
	validator.String("userId", &d.UserId).NotEmpty()
}

func AddWorkItem(ctx context.Context, client work.Client, metadata Metadata) error {
	create := newWorkCreate(metadata)
	if _, err := client.Create(ctx, create); err != nil {
		return err
	}
	return nil
}

func newWorkCreate(metadata Metadata) *work.Create {
	return &work.Create{
		Type:              processorType,
		SerialID:          pointer.FromString(metadata.UserId),
		GroupID:           pointer.FromString(NewGroupID(metadata.DataSourceId)),
		ProcessingTimeout: processingTimeoutSeconds,
		Metadata:          fromMetadata(metadata),
	}
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
	data, err := toMetadata(wrk)
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

func toMetadata(wrk *work.Work) (*Metadata, error) {
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
	if dataSourceState, ok := wrk.Metadata["dataSourceState"].(string); ok {
		data.DataSourceState = dataSourceState
	} else {
		return nil, fmt.Errorf(`expected field "dataSourceState" to exist and be a string, received %T`, wrk.Metadata["dataSourceState"])
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

func fromMetadata(data Metadata) map[string]any {
	return map[string]any{
		"userId":            data.UserId,
		"providerName":      data.ProviderName,
		"dataSourceState":   data.DataSourceState,
		"patientName":       data.PatientName,
		"restrictedTokenId": data.RestrictedTokenId,
		"emailTemplate":     data.EmailTemplate,
	}
}
