package deviceissues

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

// Metadata is intentionally empty; the device issues work item carries no
// per-run state. It exists to satisfy the generic base processor.
type Metadata struct{}

func (m *Metadata) Parse(parser structure.ObjectParser) {}

func (m *Metadata) Validate(validator structure.Validator) {}

type Processor struct {
	*workBase.Processor[Metadata]
	Dependencies
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
		Processor:    base,
		Dependencies: dependencies,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.updateDeviceIssues,
	).Process(p.Pending)
}

func (p *Processor) updateDeviceIssues() *work.ProcessResult {
	if err := p.UpdateDeviceIssues(p.Context()); err != nil {
		return p.Failing(errors.Wrap(err, "unable to update device issues"))
	}
	log.LoggerFromContext(p.Context()).Info("updated device issues")
	return nil
}
