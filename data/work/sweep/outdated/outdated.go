package outdated

import (
	"context"
	"time"

	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	summaryStore "github.com/tidepool-org/platform/summary/store"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

//go:generate mockgen -source=outdated.go -destination=test/outdated_mocks.go -package=test -typed

const (
	Type              = "org.tidepool.data.sweep.outdated"
	Quantity          = 1
	Frequency         = 1 * time.Minute
	ProcessingTimeout = 5 * time.Minute

	// PendingAvailableDuration is how long the sweep waits before running again
	PendingAvailableDuration = 1 * time.Minute

	// BatchSize is how many summaries are reported per request, and PageLimit how many requests are
	// made per run, so that one run cannot hold a processor indefinitely against a large backlog
	BatchSize = 500
	PageLimit = 20

	FailingRetryDuration       = 1 * time.Minute
	FailingRetryDurationJitter = 5 * time.Second
)

// Summaries is satisfied by summary/store.TypelessSummaries
type Summaries interface {
	ListOutdated(ctx context.Context, limit int) ([]summaryStore.OutdatedSummary, error)
	ClearOutdated(ctx context.Context, userID string, typ string, observed time.Time) error
}

type Dependencies struct {
	workBase.Dependencies
	Summaries
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.Summaries == nil {
		return errors.New("summaries is missing")
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

func NewWorkCreate() (*work.Create, error) {
	return &work.Create{
		Type:              Type,
		DeduplicationID:   pointer.From(work.DeduplicationIDSingleton),
		ProcessingTimeout: int(ProcessingTimeout.Seconds()),
	}, nil
}

type Processor struct {
	*workBase.ProcessorWithoutMetadata
	Summaries
}

func NewProcessor(dependencies Dependencies) (*Processor, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	processResultBuilder := &workBase.ProcessResultBuilder{
		ProcessResultPendingBuilder: &workBase.ConstantProcessResultPendingBuilder{
			Duration: PendingAvailableDuration,
		},
		ProcessResultFailingBuilder: &workBase.ConstantProcessResultFailingBuilder{
			Duration: FailingRetryDuration,
		},
	}

	processor, err := workBase.NewProcessorWithoutMetadata(dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	return &Processor{
		ProcessorWithoutMetadata: processor,
		Summaries:                dependencies.Summaries,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.sweep,
	).Process(p.Pending)
}

// sweep reports the summaries marked outdated by the retired mechanism as work,
// so that they are calculated once the mechanism that marked them is gone.
//
// The work is created before the mark is cleared, so that a failure between the two reports the change
// twice rather than not at all. Clearing is guarded upon the time observed, so a mark made in between
// is reported by the next run rather than discarded.
func (p *Processor) sweep() *work.ProcessResult {
	var swept int

	for range PageLimit {
		outdated, err := p.ListOutdated(p.Context(), BatchSize)
		if err != nil {
			return p.Failing(errors.Wrap(err, "unable to list outdated summaries"))
		}
		if len(outdated) == 0 {
			break
		}

		for _, summary := range outdated {
			if err = dataWorkPostprocess.Enqueue(p.Context(), p.WorkClient(), summary.UserID, dataWorkPostprocess.ReasonDataAdded); err != nil {
				return p.Failing(errors.Wrap(err, "unable to report the change"))
			}
			if err = p.ClearOutdated(p.Context(), summary.UserID, summary.Type, summary.OutdatedSince); err != nil {
				return p.Failing(errors.Wrap(err, "unable to clear the outdated summary"))
			}
		}
		swept += len(outdated)
	}

	if swept > 0 {
		log.LoggerFromContext(p.Context()).WithField("count", swept).Info("reported outdated summaries as work")
	}

	return nil
}
