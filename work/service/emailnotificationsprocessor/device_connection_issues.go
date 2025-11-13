package emailnotificationsprocessor

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/go-common/events"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
)

const (
	deviceConnectionIssuesProcessorType = "org.tidepool.processors.device.connection.issues"
)

// DeviceConnectionIssuesData is the metadata added to a [work.Work] item for notifying users about a device connection issue, the fields will be filled out by clinic-worker when it responds to CDC events on DataSources that change the DataSource.State property.
type DeviceConnectionIssuesData struct {
	UserId            string
	ProviderName      string
	RestrictedTokenId string
	PatientName       string
	DataSourceState   string
	EmailTemplate     string
}

type deviceConnectionIssuesProcessor struct {
	dependencies Dependencies
}

func newDeviceConnectionIssuesProcessor(dependencies Dependencies) *deviceConnectionIssuesProcessor {
	return &deviceConnectionIssuesProcessor{
		dependencies: dependencies,
	}
}

func (p *deviceConnectionIssuesProcessor) Type() string {
	return deviceConnectionIssuesProcessorType
}

func (p *deviceConnectionIssuesProcessor) Quantity() int {
	return Quantity
}

func (p *deviceConnectionIssuesProcessor) Frequency() time.Duration {
	return Frequency
}

func (p *deviceConnectionIssuesProcessor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) work.ProcessResult {
	data, err := toDeviceConnectionIssuesData(wrk)
	if err != nil {
		return NewFailedResult(err, wrk)
	}

	user, err := p.dependencies.Users.Get(ctx, data.UserId)
	if err != nil {
		return NewFailedResult(err, wrk)
	}
	if user == nil || user.Username == nil {
		return NewFailedResult(fmt.Errorf(`unable to find user for userId "%s"`, data.UserId), wrk)
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
		return NewFailedResult(err, wrk)
	}
	return *work.NewProcessResultDelete()
}

func (d *DeviceConnectionIssuesData) Parse(parser structure.ObjectParser) {
	d.UserId = *parser.String("userId")
	d.ProviderName = *parser.String("providerName")
	d.RestrictedTokenId = *parser.String("restrictedTokenId")
	d.PatientName = *parser.String("patientName")
	d.DataSourceState = *parser.String("dataSourceState")
	d.EmailTemplate = *parser.String("emailTemplate")
}

func (d *DeviceConnectionIssuesData) Validate(validator structure.Validator) {
	validator.String("userId", &d.UserId).NotEmpty()
	validator.String("providerName", &d.ProviderName).NotEmpty()
	validator.String("restrictedTokenId", &d.RestrictedTokenId).NotEmpty()
	validator.String("patientName", &d.PatientName).NotEmpty()
	validator.String("dataSourceState", &d.DataSourceState).NotEmpty()
	validator.String("emailTemplate", &d.EmailTemplate).NotEmpty()
}

func NewDeviceConnectionIssuesWorkCreate(metadata DeviceConnectionIssuesData) *work.Create {
	return &work.Create{
		Type:              deviceConnectionIssuesProcessorType,
		SerialID:          pointer.FromString(fmt.Sprintf("device.connection.issues.%s.%s", metadata.UserId, metadata.ProviderName)),
		GroupID:           pointer.FromString(fmt.Sprintf("device.connection.issues.%s", metadata.UserId)),
		ProcessingTimeout: ProcessingTimeoutSeconds,
		Metadata:          fromDeviceConnectionIssuesData(metadata),
	}
}

func toDeviceConnectionIssuesData(wrk *work.Work) (*DeviceConnectionIssuesData, error) {
	wrk.EnsureMetadata()
	var data DeviceConnectionIssuesData
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

func fromDeviceConnectionIssuesData(data DeviceConnectionIssuesData) map[string]any {
	return map[string]any{
		"userId":            data.UserId,
		"providerName":      data.ProviderName,
		"dataSourceState":   data.DataSourceState,
		"patientName":       data.PatientName,
		"restrictedTokenId": data.RestrictedTokenId,
		"emailTemplate":     data.EmailTemplate,
	}
}
