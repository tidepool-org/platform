package connectionissues

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	// PendingAvailableDuration is how long after a successful run the next run becomes
	// available, that is how often the clinic's connection issues are recomputed.
	PendingAvailableDuration    = 5 * time.Minute
	FailingRetryDuration        = 1 * time.Minute
	FailingRetryDurationJitter  = 10 * time.Second
	FailingRetryDurationMaximum = 15 * time.Minute
)

type Processor struct {
	*workBase.ProcessorWithoutMetadata
	ClinicClient
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
			Duration:        FailingRetryDuration,
			DurationJitter:  FailingRetryDurationJitter,
			DurationMaximum: pointer.From(FailingRetryDurationMaximum),
		},
	}

	processor, err := workBase.NewProcessorWithoutMetadata(dependencies.Dependencies,
		processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	return &Processor{
		ProcessorWithoutMetadata: processor,
		ClinicClient:             dependencies.ClinicClient,
	}, nil
}

// Process asks the clinic to update its connection issues and then schedules the next
// run. A failure is retried with backoff and never drops the singleton work item.
func (p *Processor) Process(ctx context.Context, wrk *work.Work,
	processingUpdater work.ProcessingUpdater) *work.ProcessResult {

	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.updateConnectionIssues,
	).Process(p.Pending)
}

func (p *Processor) updateConnectionIssues() *work.ProcessResult {
	if err := p.ClinicClient.UpdateConnectionIssues(p.Context()); err != nil {
		return p.Failing(errors.Wrap(err, "unable to update connection issues"))
	}
	return nil
}
