package emailnotificationsprocessor

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/go-common/events"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"

	. "github.com/tidepool-org/platform/work/service/emailnotificationsprocessor/metadata"
)

const (
	deviceConnectionIssuesProcessorType = "org.tidepool.processors.device.connection.issues"
)

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
