package subscribe

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	PendingAvailableDuration   = 24 * time.Hour
	FailingRetryDuration       = 10 * time.Minute
	FailingRetryDurationJitter = time.Minute
)

type Processor struct {
	*workBase.ProcessorWithoutMetadata
	OuraClient
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

	processorWithoutMetadata, err := workBase.NewProcessorWithoutMetadata(dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	return &Processor{
		ProcessorWithoutMetadata: processorWithoutMetadata,
		OuraClient:               dependencies.OuraClient,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.synchronizeSubscriptions,
	).Process(p.Pending)
}

func (p *Processor) synchronizeSubscriptions() *work.ProcessResult {
	// TODO: Implement
	return nil
}
