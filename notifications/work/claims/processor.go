package claims

import (
	"context"
	"time"

	clinicClient "github.com/tidepool-org/clinic/client"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	notificationsHistory "github.com/tidepool-org/platform/notifications/history"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	FailingRetryDuration       = 1 * time.Minute
	FailingRetryDurationJitter = 5 * time.Second
)

type Metadata struct {
	ClinicID   string    `json:"clinicId,omitempty"`
	UserID     string    `json:"userId,omitempty"`
	WhenToSend time.Time `json:"whenToSend,omitzero"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("clinicId"); ptr != nil {
		m.ClinicID = *ptr
	}
	if ptr := parser.String("userId"); ptr != nil {
		m.UserID = *ptr
	}
	if ptr := parser.Time("whenToSend", time.RFC3339Nano); ptr != nil {
		m.WhenToSend = *ptr
	}
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String("clinicId", &m.ClinicID).NotEmpty()
	validator.String("userId", &m.UserID).Using(user.IDValidator)
}

type Processor struct {
	*workBase.Processor[Metadata]
	Dependencies
	patient *clinicClient.Patient
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

	return &Processor{
		Processor:    processor,
		Dependencies: dependencies,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.fetchPatient,
		p.process,
	).Process(p.Delete)
}

func (p *Processor) fetchPatient() *work.ProcessResult {
	userID := p.Metadata().UserID

	patient, err := p.GetPatient(p.Context(), p.Metadata().ClinicID, userID)
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to get patient"))
	} else if patient == nil {
		return p.Failed(errors.New("patient is missing"))
	} else if patient.Email == nil || *patient.Email == "" {
		return p.Failed(errors.New("patient email is missing"))
	}
	p.patient = patient

	return nil
}

func (p *Processor) process() *work.ProcessResult {
	// If user already claimed they will no longer have the custodian field set
	if p.patient.Permissions == nil || p.patient.Permissions.Custodian == nil {
		p.recordHistoryEntry(notificationsHistory.NotificationConditionsExpired)
		return nil
	}

	p.recordHistoryEntry(notificationsHistory.NotificationAttempted)

	if _, err := p.ResendAccountSignupWithResponse(p.Context(), *p.patient.Email); err != nil {
		p.recordHistoryEntryWithError(notificationsHistory.NotificationEmailError, errors.Wrap(err, "unable to send email for account claim"))
		return p.Failing(errors.Wrap(err, "unable to send email for account claim"))
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
		GroupID:       NewGroupID(p.Metadata().UserID),
		Metadata:      p.MetadataEncoded(),
		UserID:        p.Metadata().UserID,
		Email:         pointer.Default(p.patient.Email, ""),
		Error:         err,
	}
	if err := p.Create(p.Context(), entry); err != nil {
		log.LoggerFromContext(p.Context()).WithError(err).Error("unable to record history entry")
	}
}
