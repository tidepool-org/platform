package work

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oura/shopify"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type      = "org.tidepool.processors.oura.shopify.reconcile"
	Quantity  = 1
	Frequency = time.Minute

	PendingRetryDuration       = 30 * time.Minute
	FailingRetryDurationJitter = 10 * time.Second
	FailingRetryDuration       = ProcessingTimeout * 2
	ProcessingTimeout          = 3 * time.Minute

	MetadataKeyUpdatedSince = "updatedSince"
)

type Dependencies struct {
	OrderProcessor *shopify.OrderProcessor
}

func (d Dependencies) Validate() error {
	if d.OrderProcessor == nil {
		return errors.New("order processor is missing")
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
		Type:              Type,
		DeduplicationID:   pointer.FromString(Type),
		ProcessingTimeout: int(ProcessingTimeout.Seconds()),
		Metadata: map[string]any{
			MetadataKeyUpdatedSince: time.Now(),
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
	updatedSince, err := p.updatedSinceFromMetadata()
	if err != nil {
		return p.Failed(err)
	} else if updatedSince == nil || updatedSince.IsZero() {
		return p.Failed(errors.New("updated since is missing"))
	}

	latestUpdatedTime, err := p.OrderProcessor.ReconcileUpdatedOrders(p.Context(), *updatedSince)
	p.Work().Metadata[MetadataKeyUpdatedSince] = latestUpdatedTime

	if err != nil {
		return p.Failing(err)
	}

	p.Logger().Info("reconciled orders")
	return nil
}

func (p *Processor) updatedSinceFromMetadata() (*time.Time, error) {
	parser := p.MetadataParser()
	updatedSince := parser.Time(MetadataKeyUpdatedSince, time.RFC3339)
	return updatedSince, parser.Error()
}
