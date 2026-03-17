package work

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura/jotform"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type      = "org.tidepool.processors.oura.jotform.reconcile"
	Quantity  = 1
	Frequency = time.Minute

	PendingRetryDuration       = 30 * time.Minute
	FailingRetryDurationJitter = 10 * time.Second
	FailingRetryDuration       = ProcessingTimeout * 2
	ProcessingTimeout          = 3 * time.Minute

	MetadataKeyLastProcessedSubmissionID = "lastProcessedSubmissionId"
	initialSubmissionID                  = "0"
)

type Metadata struct {
	LastProcessedSubmissionID *string `json:"lastProcessedSubmissionId,omitempty" bson:"lastProcessedSubmissionId,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.LastProcessedSubmissionID = parser.String(MetadataKeyLastProcessedSubmissionID)
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String(MetadataKeyLastProcessedSubmissionID, m.LastProcessedSubmissionID).Exists().NotEmpty()
}

type Dependencies struct {
	workBase.Dependencies
	SubmissionProcessor *jotform.SubmissionProcessor
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.SubmissionProcessor == nil {
		return errors.New("submission processor is missing")
	}
	return nil
}

func NewProcessorFactory(dependencies Dependencies) (*workBase.ProcessorFactory, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}
	processorFactory := func() (work.Processor, error) { return NewProcessor(dependencies) }
	return workBase.NewProcessorFactory(Type, Quantity, Frequency, processorFactory)
}

func NewProcessor(dependencies Dependencies) (*Processor, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	processResultBuilder := &workBase.ProcessResultBuilder{
		ProcessResultPendingBuilder: &workBase.ConstantProcessResultPendingBuilder{
			Duration: PendingRetryDuration,
		},
		ProcessResultFailingBuilder: &workBase.ExponentialProcessResultFailingBuilder{
			Duration:       FailingRetryDuration,
			DurationJitter: FailingRetryDurationJitter,
		},
	}

	base, err := workBase.NewProcessor[Metadata](dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	return &Processor{
		Dependencies: dependencies,
		Processor:    base,
	}, nil
}

func EnsureReconcilerWorkItemExists(ctx context.Context, client work.Client) error {
	create := &work.Create{
		Type:              Type,
		DeduplicationID:   pointer.FromString(Type),
		ProcessingTimeout: int(ProcessingTimeout.Seconds()),
		Metadata: map[string]any{
			MetadataKeyLastProcessedSubmissionID: initialSubmissionID,
		},
	}
	if _, err := client.Create(ctx, create); err != nil {
		return err
	}
	return nil
}

type Processor struct {
	*workBase.Processor[Metadata]
	Dependencies
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.reconcile,
	).Process(p.Pending)
}

func (p *Processor) reconcile() *work.ProcessResult {
	if p.Metadata().LastProcessedSubmissionID == nil {
		return p.Failed(errors.New("last processed submission id is missing"))
	}

	result, err := p.SubmissionProcessor.Reconcile(p.Context(), *p.Metadata().LastProcessedSubmissionID)
	p.Metadata().LastProcessedSubmissionID = pointer.FromString(result.LastProcessedID)
	p.AddFieldsToContext(log.Fields{
		"processed": result.TotalProcessed,
	})

	if err != nil {
		return p.Failing(err)
	}
	log.LoggerFromContext(p.Context()).Info("reconciled submissions")
	return nil
}
