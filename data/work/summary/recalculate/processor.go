package recalculate

import (
	"context"
	"time"

	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	// BatchSize is the number of users whose recalculation is requested each time the work
	// is processed
	BatchSize = 250

	// PendingMaximum is the number of postprocess work items pending at which no more
	// recalculations are requested, so that they don't flood the postprocess work queue
	PendingMaximum = 500

	PendingAvailableDuration    = 30 * time.Second
	FailingRetryDuration        = 1 * time.Minute
	FailingRetryDurationJitter  = 5 * time.Second
	FailingRetryDurationMaximum = 1 * time.Hour
)

const MetadataKeyLastUserID = "lastUserId"

type Metadata struct {
	// LastUserID is the last user whose recalculation was requested
	LastUserID *string `json:"lastUserId,omitempty" bson:"lastUserId,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.LastUserID = parser.String(MetadataKeyLastUserID)
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String(MetadataKeyLastUserID, m.LastUserID).NotEmpty()
}

// Processor requests the recalculation of the summaries of every user with a summary, a
// batch of users at a time, by creating postprocess work for each user. Postprocess work is
// used, rather than recalculating the summaries directly, so that the recalculation is
// serialized with, and coalesced into, any other changes to the data of the user, and the
// clinic service is notified of the recalculated summaries.
type Processor struct {
	*workBase.Processor[Metadata]
	SummaryLister
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

	processor, err := workBase.NewProcessor[Metadata](dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	return &Processor{
		Processor:     processor,
		SummaryLister: dependencies.SummaryLister,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.awaitPostprocess,
		p.requestRecalculations,
	).Process(p.Pending)
}

// awaitPostprocess defers the work while the postprocess work queue is backed up
func (p *Processor) awaitPostprocess() *work.ProcessResult {
	filter := &work.Filter{
		Types: pointer.From([]string{dataWorkPostprocess.Type}),
		State: pointer.From(work.StatePending),
	}
	pagination := &page.Pagination{Page: 0, Size: PendingMaximum}

	wrks, err := p.WorkClient().List(p.Context(), filter, pagination)
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to list pending postprocess work"))
	}
	if len(wrks) >= PendingMaximum {
		log.LoggerFromContext(p.Context()).
			Debug("postprocess work is backed up, deferring summary recalculation")
		return p.Pending()
	}

	return nil
}

// requestRecalculations requests the recalculation of the summaries of the next batch of
// users, and succeeds once there are no more users
//
// The work succeeds rather than being deleted, so that it remains to deduplicate the work
// created by any later start of the service.
func (p *Processor) requestRecalculations() *work.ProcessResult {
	userIDs, err := p.ListUserIDs(p.Context(), p.Metadata().LastUserID, BatchSize)
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to list user ids"))
	}
	if len(userIDs) == 0 {
		log.LoggerFromContext(p.Context()).Info("requested the recalculation of every summary")
		return p.Success()
	}

	// If this fails partway, the users already requested are requested again on retry,
	// which postprocess coalesces with the work already pending for the user
	for _, userID := range userIDs {
		err = dataWorkPostprocess.Enqueue(p.Context(), p.WorkClient(), userID,
			dataWorkPostprocess.ReasonSummaryRecalculation)
		if err != nil {
			return p.Failing(errors.Wrap(err, "unable to request summary recalculation"))
		}
	}
	p.Metadata().LastUserID = pointer.From(userIDs[len(userIDs)-1])

	log.LoggerFromContext(p.Context()).WithField("count", len(userIDs)).
		Debug("requested summary recalculations")

	return nil
}
