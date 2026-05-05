package claims

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
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
		p.process,
	).Process(p.Delete)
}

func (p *Processor) process() *work.ProcessResult {
	patient, err := p.GetPatient(p.Context(), p.Metadata().ClinicID, p.Metadata().UserID)
	if err != nil {
		return p.Failing(err)
	} else if patient == nil {
		return p.Failing(errors.Newf("unable to find patient with user id %q", p.Metadata().UserID))
	} else if patient.Email == nil || *patient.Email == "" {
		return p.Failing(errors.Newf("unable to find email for patient with user id %q", p.Metadata().UserID))
	}

	// If user already claimed they will no longer have the custodian field set
	if patient.Permissions == nil || patient.Permissions.Custodian == nil {
		return nil
	}

	if _, err := p.ResendAccountSignupWithResponse(p.Context(), *patient.Email); err != nil {
		return p.Failing(errors.New("unable to resend account signup email"))
	}

	return nil
}
