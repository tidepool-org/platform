package requests

import (
	"context"
	"time"

	"github.com/tidepool-org/go-common/events"

	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	userWork "github.com/tidepool-org/platform/user/work"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	FailingRetryDuration       = 1 * time.Minute
	FailingRetryDurationJitter = 5 * time.Second
)

type Metadata struct {
	ClinicID          string    `json:"clinicId,omitempty"`
	Email             string    `json:"email,omitempty"`
	EmailTemplate     string    `json:"emailTemplate,omitempty"`
	PatientName       string    `json:"patientName,omitempty"`
	ProviderName      string    `json:"providerName,omitempty"`
	RestrictedTokenID string    `json:"restrictedTokenId,omitempty"`
	UserID            string    `json:"userId,omitempty"`
	WhenToSend        time.Time `json:"whenToSend,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("clinicId"); ptr != nil {
		m.ClinicID = *ptr
	}
	if ptr := parser.String("email"); ptr != nil {
		m.Email = *ptr
	}
	if ptr := parser.String("emailTemplate"); ptr != nil {
		m.EmailTemplate = *ptr
	}
	if ptr := parser.String("patientName"); ptr != nil {
		m.PatientName = *ptr
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
	if ptr := parser.Time("whenToSend", time.RFC3339Nano); ptr != nil {
		m.WhenToSend = *ptr
	}
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String("clinicId", &m.ClinicID).NotEmpty()
	validator.String("email", &m.Email).NotEmpty()
	validator.String("emailTemplate", &m.EmailTemplate).NotEmpty()
	validator.String("patientName", &m.PatientName).NotEmpty()
	validator.String("providerName", &m.ProviderName).NotEmpty()
	validator.String("restrictedTokenId", &m.RestrictedTokenID).NotEmpty()
	validator.String("userId", &m.UserID).NotEmpty()
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
		p.process,
	).Process(p.Delete)
}

func (p *Processor) process() *work.ProcessResult {
	if result := p.FetchUser(p.Metadata().UserID); result != nil {
		return result
	}

	username := p.User().Username
	if username == nil {
		return p.Failed(errors.New("user username is missing"))
	}

	filter := &dataSource.Filter{
		ProviderName: pointer.FromString(p.Metadata().ProviderName),
		State:        pointer.FromString(dataSource.StateConnected),
	}
	dataSrcs, err := p.DataSourceClient.List(p.Context(), *p.User().UserID, filter, page.NewPaginationMinimum())
	if err != nil {
		return p.Failing(err)
	} else if len(dataSrcs) > 0 {
		return nil // User now has a connected dataSource so no email to send
	}

	var clinicName string
	clinic, err := p.ClinicClient.GetClinic(p.Context(), p.Metadata().ClinicID)
	if err != nil {
		return p.Failing(errors.Wrapf(err, `error getting clinic`))
	} else if clinic != nil {
		clinicName = clinic.Name
	}

	variables := map[string]string{
		"ClinicName":        clinicName,
		"PatientName":       p.Metadata().PatientName,
		"ProviderName":      p.Metadata().ProviderName,
		"RestrictedTokenId": p.Metadata().RestrictedTokenID,
	}
	templateEvent := events.SendEmailTemplateEvent{
		Recipient: *username,
		Template:  p.Metadata().EmailTemplate,
		Variables: variables,
	}
	if err := p.MailerClient.SendEmailTemplate(p.Context(), templateEvent); err != nil {
		return p.Failing(err)
	}

	return nil
}
