package work

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/oura/shopify"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
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
)

const (
	MetadataKeyUpdatedSince = "updatedSince"
)

type Metadata struct {
	UpdatedSince time.Time `json:"updatedSince,omitzero"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	if ptr := parser.Time(MetadataKeyUpdatedSince, time.RFC3339); ptr != nil {
		m.UpdatedSince = *ptr
	}
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.Time(MetadataKeyUpdatedSince, &m.UpdatedSince).NotZero()
}

type Dependencies struct {
	workBase.Dependencies
	OrderProcessor *shopify.OrderProcessor
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
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

type Processor struct {
	*workBase.Processor[Metadata]
	Dependencies
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

	base, err := workBase.NewProcessor[Metadata](dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	return &Processor{
		Dependencies: dependencies,
		Processor:    base,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.reconcile,
	).Process(p.Pending)
}

func (p *Processor) reconcile() *work.ProcessResult {
	latestUpdatedTime, err := p.OrderProcessor.ReconcileUpdatedOrders(p.Context(), p.Metadata().UpdatedSince)
	p.Metadata().UpdatedSince = latestUpdatedTime

	if err != nil {
		return p.Failing(err)
	}

	log.LoggerFromContext(p.Context()).Info("reconciled orders")
	return nil
}

func EnsureReconcilerWorkItemExists(ctx context.Context, client work.Client) error {
	if client == nil {
		return errors.New("client is missing")
	}
	create, err := metadata.WithMetadata(
		&work.Create{
			Type:              Type,
			DeduplicationID:   pointer.FromString(work.DeduplicationIDSingleton),
			ProcessingTimeout: int(ProcessingTimeout.Seconds()),
		},
		&Metadata{
			UpdatedSince: time.Now(),
		},
	)
	if err != nil {
		return errors.Wrap(err, "unable to create work create")
	} else if _, err := client.Create(ctx, create); err != nil {
		return err
	}
	return nil
}
