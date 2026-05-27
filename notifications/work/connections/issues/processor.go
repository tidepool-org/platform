package issues

import (
	"context"
	"time"

	"github.com/tidepool-org/go-common/events"

	"github.com/tidepool-org/platform/auth"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	notificationsHistory "github.com/tidepool-org/platform/notifications/history"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/user"
	userWork "github.com/tidepool-org/platform/user/work"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	FailingRetryDuration       = 1 * time.Minute
	FailingRetryDurationJitter = 5 * time.Second
)

type Metadata struct {
	DataSourceState   string `json:"dataSourceState,omitempty"`
	DataSourceID      string `json:"dataSourceId,omitempty"`
	EmailTemplate     string `json:"emailTemplate,omitempty"`
	FullName          string `json:"fullName,omitempty"`
	ProviderName      string `json:"providerName,omitempty"`
	RestrictedTokenID string `json:"restrictedTokenId,omitempty"`
	UserID            string `json:"userId,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("dataSourceState"); ptr != nil {
		m.DataSourceState = *ptr
	}
	if ptr := parser.String("dataSourceId"); ptr != nil {
		m.DataSourceID = *ptr
	}
	if ptr := parser.String("emailTemplate"); ptr != nil {
		m.EmailTemplate = *ptr
	}
	if ptr := parser.String("fullName"); ptr != nil {
		m.FullName = *ptr
	}
	if ptr := parser.String("providerName"); ptr != nil {
		m.ProviderName = *ptr
	}
	if ptr := parser.String("restrictedTokenId"); ptr != nil {
		m.RestrictedTokenID = *ptr
	}
	if ptr := parser.String("userId"); ptr != nil {
		m.UserID = *ptr
	}
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String("dataSourceState", &m.DataSourceState).OneOf(dataSource.States()...)
	validator.String("dataSourceId", &m.DataSourceID).NotEmpty() // NOTE: _id (not id, as expected)
	validator.String("emailTemplate", &m.EmailTemplate).NotEmpty()
	validator.String("providerName", &m.ProviderName).Using(auth.ProviderNameValidator)
	validator.String("restrictedTokenId", &m.RestrictedTokenID).Using(auth.RestrictedTokenIDValidator)
	validator.String("userId", &m.UserID).Using(user.IDValidator)
}

type UserMixin = userWork.Mixin

type Processor struct {
	*workBase.Processor[Metadata]
	UserMixin
	Dependencies
}

func NewProcessor(dependencies Dependencies) (*Processor, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	processResultBuilder := &workBase.ProcessResultBuilder{
		ProcessResultFailingBuilder: &workBase.ExponentialProcessResultFailingBuilder{
			Duration:       FailingRetryDuration,
			DurationJitter: FailingRetryDurationJitter,
		},
	}

	processor, err := workBase.NewProcessor[Metadata](dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}
	userMixin, err := userWork.NewMixin(processor, dependencies.UserClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create user mixin")
	}

	return &Processor{
		Processor:    processor,
		UserMixin:    userMixin,
		Dependencies: dependencies,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.fetchUser,
		p.process,
	).Process(p.Delete)
}

func (p *Processor) fetchUser() *work.ProcessResult {
	return p.FetchUser(p.Metadata().UserID)
}

func (p *Processor) process() *work.ProcessResult {
	if p.User().Username == nil {
		p.recordHistoryEntryWithError(notificationsHistory.NotificationGeneralError, errors.New("user email is missing"))
		return p.Failed(errors.New("user email is missing"))
	}

	p.recordHistoryEntry(notificationsHistory.NotificationAttempted)

	templateEvent := events.SendEmailTemplateEvent{
		Recipient: *p.User().Username,
		Template:  p.Metadata().EmailTemplate,
		Variables: map[string]string{
			"RestrictedTokenId": p.Metadata().RestrictedTokenID,
			"FullName":          p.Metadata().FullName,
			"ProviderName":      p.Metadata().ProviderName,
		},
	}
	if err := p.SendEmailTemplate(p.Context(), templateEvent); err != nil {
		p.recordHistoryEntryWithError(notificationsHistory.NotificationEmailError, errors.Wrap(err, "unable to send email for connection issue"))
		return p.Failing(errors.Wrap(err, "unable to send email for connection issue"))
	}

	p.recordHistoryEntry(notificationsHistory.NotificationEmailSent)
	return nil
}

func (p *Processor) recordHistoryEntry(eventType string) {
	p.recordHistoryEntryWithError(eventType, nil)
}

func (p *Processor) recordHistoryEntryWithError(eventType string, err error) {
	entry := notificationsHistory.Entry{
		EventType:     eventType,
		ProcessorType: Type,
		GroupID:       NewGroupID(p.Metadata().DataSourceID),
		Metadata:      p.MetadataEncoded(),
		UserID:        p.Metadata().UserID,
		DataSourceID:  p.Metadata().DataSourceID,
		Email:         pointer.Default(p.User().Username, ""),
		Error:         err,
	}
	if err := p.Create(p.Context(), entry); err != nil {
		log.LoggerFromContext(p.Context()).WithError(err).Error("unable to record history entry")
	}
}
