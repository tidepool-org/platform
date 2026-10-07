package recalculate

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

//go:generate mockgen -source=factory.go -destination=test/factory_mocks.go -package=test -typed

const (
	Type              = "org.tidepool.data.summary.recalculate"
	Quantity          = 1
	Frequency         = 30 * time.Second
	ProcessingTimeout = 5 * time.Minute
)

// ID identifies the recalculation of every summary that is required by the current version
// of the code. It is the deduplication id of the work, which is created when the data
// service starts. On success, the work is kept rather than deleted, so that the work
// created by any later start with the same id is discarded as a duplicate, and so the
// recalculation runs exactly once.
//
// Change it to a new, never before used, value whenever every summary must be recalculated,
// such as when the glucose ranges are modified. To run the recalculation with the current
// id again, delete its work.
const ID = "BACK-4158-gap-based-cut-points"

type SummaryLister interface {
	// ListUserIDs returns, in ascending order, the ids of up to limit users with a summary,
	// starting after the given user id, if any.
	ListUserIDs(ctx context.Context, afterUserID *string, limit int) ([]string, error)
}

type Dependencies struct {
	workBase.Dependencies
	SummaryLister
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.SummaryLister == nil {
		return errors.New("summary lister is missing")
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

// NewWorkCreate returns the work to recalculate every summary once for the given id
func NewWorkCreate(id string) (*work.Create, error) {
	if id == "" {
		return nil, errors.New("id is missing")
	}
	return &work.Create{
		Type:                    Type,
		DeduplicationID:         pointer.From(id),
		ProcessingAvailableTime: time.Now(),
		ProcessingTimeout:       int(ProcessingTimeout.Seconds()),
	}, nil
}
