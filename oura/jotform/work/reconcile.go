package work

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura/jotform"
	"github.com/tidepool-org/platform/pointer"
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
	reconcilerWorkID                     = "reconciler"
)

type Dependencies struct {
	SubmissionProcessor *jotform.SubmissionProcessor
}

func (d Dependencies) Validate() error {
	if d.SubmissionProcessor == nil {
		return errors.New("submission processor is missing")
	}
	return nil
}

func NewProcessorFactory(dependencies Dependencies) (*workBase.ProcessorFactory, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies are invalid")
	}
	processorFactory := func() (work.Processor, error) { return NewProcessor(dependencies) }
	return workBase.NewProcessorFactory(Type, Quantity, Frequency, processorFactory)
}

func NewProcessor(dependencies Dependencies) (*Processor, error) {
	processResultBuilder := &workBase.ProcessResultBuilder{
		ProcessResultPendingBuilder: &workBase.ConstantProcessResultPendingBuilder{
			Duration: PendingRetryDuration,
		},
		ProcessResultFailingBuilder: &workBase.ExponentialProcessResultFailingBuilder{
			Duration:       FailingRetryDuration,
			DurationJitter: FailingRetryDurationJitter,
		},
	}

	base, err := workBase.NewProcessor(processResultBuilder)
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
		Type:                    Type,
		DeduplicationID:         pointer.FromString(reconcilerWorkID),
		ProcessingTimeout:       int(ProcessingTimeout.Seconds()),
		ProcessingAvailableTime: time.Now(),
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
	*workBase.Processor
	Dependencies
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) *work.ProcessResult {
	return work.ProcessPipeline{
		p.ProcessPipelineFunc(ctx, wrk, updater),
		p.reconcile,
		p.Pending,
	}.Process()
}

func (p *Processor) reconcile() *work.ProcessResult {
	result, err := p.SubmissionProcessor.Reconcile(p.Context(), p.lastProcessedSubmissionIDFromMetadata())
	p.Work().Metadata[MetadataKeyLastProcessedSubmissionID] = result.LastProcessedID
	p.AddFieldsToContext(log.Fields{
		"processed": result.TotalProcessed,
		"errors":    result.TotalErrors,
	})

	if err != nil {
		return p.Failing(err)
	}

	p.Logger().Info("reconciled submissions")
	return nil
}

func (p *Processor) lastProcessedSubmissionIDFromMetadata() string {
	parser := p.MetadataParser()
	return pointer.Default(parser.String(MetadataKeyLastProcessedSubmissionID), initialSubmissionID)
}
