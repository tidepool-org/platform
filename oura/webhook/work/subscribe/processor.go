package subscribe

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type      = "org.tidepool.oura.work.subscribe"
	Quantity  = 1
	Frequency = 5 * time.Second

	PendingAvailableDuration   = 1 * time.Hour
	FailingRetryDuration       = 1 * time.Minute
	FailingRetryDurationJitter = 5 * time.Second

	ProcessingTimeout = 60 // Seconds
)

type Dependencies struct {
	workBase.Dependencies
	Client oura.Client
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return errors.Wrap(err, "dependencies is invalid")
	}
	if d.Client == nil {
		return errors.New("client is missing")
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

type Processor struct {
	*workBase.Processor
	Client oura.Client
}

func NewProcessor(dependencies Dependencies) (*Processor, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	processResultBuilder := &workBase.ProcessResultBuilder{
		ProcessResultPendingBuilder: &workBase.ConstantProcessResultPendingBuilder{
			Duration: PendingAvailableDuration,
		},
		ProcessResultFailingBuilder: &workBase.ExponentialProcessResultFailingBuilder{
			Duration:       FailingRetryDuration,
			DurationJitter: FailingRetryDurationJitter,
		},
	}

	processor, err := workBase.NewProcessor(dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	return &Processor{
		Processor: processor,
		Client:    dependencies.Client,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		func() *work.ProcessResult { return nil }, // TODO: Implement
	).Process(p.Pending)
}

func NewWorkCreate() *work.Create {
	return &work.Create{
		Type:              Type,
		GroupID:           pointer.FromString(Type),
		DeduplicationID:   pointer.FromString(Type),
		ProcessingTimeout: ProcessingTimeout,
	}
}
