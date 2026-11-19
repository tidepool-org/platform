package recalculate

import (
	"context"
	"time"

	dataWorkSummaryRecalculateUser "github.com/tidepool-org/platform/data/work/summary/recalculate/user"
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

	// PendingMaximum is the number of user recalculations pending at which no more are
	// requested, so that they don't flood the work queue
	PendingMaximum = 500

	PendingAvailableDuration    = 30 * time.Second
	FailingRetryDuration        = 1 * time.Minute
	FailingRetryDurationJitter  = 0.1
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
// batch of users at a time, by creating work for each user. The work of a user
// recalculates their summaries and syncs them to the clinic service.
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
		func() *work.ProcessResult { return p.supersede(wrk) },
		p.awaitPending,
		p.requestRecalculations,
	).Process(p.Pending)
}

// supersede deletes the work if it isn't for the current recalculation, which is the one
// identified by [ID].
//
// When the id is changed before an earlier recalculation has requested every user, the
// work of the earlier recalculation remains. The current recalculation starts again from
// the first user, so the earlier one has nothing left to contribute, and would only
// request users a second time. It's deleted rather than kept as a success, so that it runs
// again in full if a rollback makes its id current.
func (p *Processor) supersede(wrk *work.Work) *work.ProcessResult {
	if wrk.DeduplicationID != nil && *wrk.DeduplicationID == ID {
		return nil
	}
	log.LoggerFromContext(p.Context()).WithField("currentId", ID).
		Info("summary recalculation is superseded, deleting it")
	return p.Delete()
}

// awaitPending defers the work while the user recalculations already requested are backed
// up
func (p *Processor) awaitPending() *work.ProcessResult {
	filter := &work.Filter{
		Types: pointer.From([]string{dataWorkSummaryRecalculateUser.Type}),
		State: pointer.From(work.StatePending),
	}
	pagination := &page.Pagination{Page: 0, Size: PendingMaximum}

	wrks, err := p.WorkClient().List(p.Context(), filter, pagination)
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to list pending user recalculations"))
	}
	if len(wrks) >= PendingMaximum {
		log.LoggerFromContext(p.Context()).
			Debug("user recalculations are backed up, deferring summary recalculation")
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
	// which is discarded as a duplicate of the work still pending for the user
	for _, userID := range userIDs {
		create, err := dataWorkSummaryRecalculateUser.NewWorkCreate(ID, userID)
		if err != nil {
			return p.Failing(errors.Wrap(err, "unable to request summary recalculation"))
		}
		if _, err = p.WorkClient().Create(p.Context(), create); err != nil {
			return p.Failing(errors.Wrap(err, "unable to request summary recalculation"))
		}
	}
	p.Metadata().LastUserID = pointer.From(userIDs[len(userIDs)-1])

	log.LoggerFromContext(p.Context()).WithField("count", len(userIDs)).
		Debug("requested summary recalculations")

	return nil
}
