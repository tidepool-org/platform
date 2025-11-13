package emailnotificationsprocessor

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/go-common/events"
	"github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"

	. "github.com/tidepool-org/platform/work/service/emailnotificationsprocessor/metadata"
)

const (
	connectAccountProcessorType = "org.tidepool.processors.connect.account"
)

type connectAccountProcessor struct {
	dependencies Dependencies
}

func newConnectAccountProcessor(dependencies Dependencies) *connectAccountProcessor {
	return &connectAccountProcessor{
		dependencies: dependencies,
	}
}

func (p *connectAccountProcessor) Type() string {
	return connectAccountProcessorType
}

func (p *connectAccountProcessor) Quantity() int {
	return Quantity
}

func (p *connectAccountProcessor) Frequency() time.Duration {
	return Frequency
}

func (p *connectAccountProcessor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) work.ProcessResult {
	data, err := toConnectAccountData(wrk)
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
	filter := source.NewFilter()
	filter.ProviderName = pointer.FromStringArray([]string{data.ProviderName})
	filter.State = pointer.FromStringArray([]string{"connected"})
	connectedDataSources, err := p.dependencies.DataSources.List(ctx, data.UserId, filter, nil)
	if err != nil {
		return NewFailedResult(err, wrk)
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
		return NewFailedResult(err, wrk)
	}
	return *work.NewProcessResultDelete()
}

func NewConnectAccountWorkCreate(notBefore time.Time, metadata ConnectAccountReminderData) *work.Create {
	return &work.Create{
		Type:                    connectAccountProcessorType,
		SerialID:                pointer.FromString(fmt.Sprintf("connect.%s.%s", metadata.UserId, metadata.ProviderName)),
		GroupID:                 pointer.FromString(fmt.Sprintf("connect.%s.%s", metadata.UserId, metadata.ProviderName)),
		ProcessingTimeout:       ProcessingTimeoutSeconds,
		ProcessingAvailableTime: notBefore,
		Metadata:                fromConnectAccountData(metadata),
	}
}

func toConnectAccountData(wrk *work.Work) (*ConnectAccountReminderData, error) {
	wrk.EnsureMetadata()
	var data ConnectAccountReminderData
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

func fromConnectAccountData(data ConnectAccountReminderData) map[string]any {
	return map[string]any{
		"userId":            data.UserId,
		"providerName":      data.ProviderName,
		"patientName":       data.PatientName,
		"restrictedTokenId": data.RestrictedTokenId,
		"emailTemplate":     data.EmailTemplate,
	}
}
