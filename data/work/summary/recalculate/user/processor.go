package user

import (
	"context"

	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	dataWorkSummary "github.com/tidepool-org/platform/data/work/summary"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

// A Processor recalculates the CGM and BGM summaries of a user from all of their data, and
// syncs them to the clinic service.
//
// Unlike the postprocess work, it doesn't trigger an EHR sync. EHR syncs are for new data,
// and a recalculation has none.
type Processor struct {
	*workBase.Processor[Metadata]
	Summarizers
	ClinicsClient

	summariesUpdate SummariesUpdate
}

func NewProcessor(dependencies Dependencies) (*Processor, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	processResultBuilder := &workBase.ProcessResultBuilder{
		ProcessResultFailingBuilder: &workBase.ExponentialProcessResultFailingBuilder{
			Duration:        FailingRetryDuration,
			DurationJitter:  FailingRetryDurationJitter,
			DurationMaximum: pointer.From(FailingRetryDurationMaximum),
		},
	}

	processor, err := workBase.NewProcessor[Metadata](dependencies.Dependencies,
		processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	return &Processor{
		Processor:     processor,
		Summarizers:   dependencies.Summarizers,
		ClinicsClient: dependencies.ClinicsClient,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work,
	processingUpdater work.ProcessingUpdater) *work.ProcessResult {

	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		func() *work.ProcessResult { return p.validateWork(wrk) },
		p.recalculateSummaries,
		p.syncClinicSummaries,
	).Process(p.Delete)
}

// validateWork fails if the work isn't serialized with the postprocess work of its user.
func (p *Processor) validateWork(wrk *work.Work) *work.ProcessResult {
	userID := p.Metadata().UserID
	if userID == nil {
		return p.Failed(errors.New("work is invalid; metadata user id is missing"))
	}
	serialID := dataWorkPostprocess.IDFromUserID(*userID)
	if wrk.SerialID == nil || *wrk.SerialID != serialID {
		return p.Failed(
			errors.New("work is invalid; serial id does not match metadata user id"))
	}
	return nil
}

func (p *Processor) recalculateSummaries() *work.ProcessResult {
	var err error
	p.summariesUpdate, err = p.RecalculateSummaries(p.Context(), *p.Metadata().UserID)

	// The changes made are recorded in the metadata before they are synced to the clinic
	// service, so that a failure between the two is retried. They are recorded even when
	// the recalculation fails partway, as a partial update is returned alongside the error.
	workMetadata := p.Metadata()
	changed := dataWorkSummary.RecordSummariesUpdate(&workMetadata.PendingSummaryUpdates,
		&workMetadata.PendingSummaryDeletes, p.summariesUpdate)
	if err != nil {
		return p.Failing(err)
	}
	if changed {
		if result := p.ProcessingUpdate(); result != nil {
			return result
		}
		log.LoggerFromContext(p.Context()).WithFields(log.Fields{
			"updated": p.summariesUpdate.UpdatedTypes,
			"deleted": p.summariesUpdate.Deleted,
		}).Info("recalculated user summaries")
	}

	return nil
}

func (p *Processor) syncClinicSummaries() *work.ProcessResult {
	workMetadata := p.Metadata()
	synced, err := dataWorkSummary.SyncClinicSummaries(p.Context(), p.ClinicsClient,
		*workMetadata.UserID, workMetadata.PendingSummaryUpdates,
		workMetadata.PendingSummaryDeletes, p.summariesUpdate)
	if err != nil {
		return p.Failing(err)
	}
	if synced {
		log.LoggerFromContext(p.Context()).WithFields(log.Fields{
			"updated": workMetadata.PendingSummaryUpdates,
			"deleted": workMetadata.PendingSummaryDeletes,
		}).Debug("synced user summaries with the clinic service")
	}

	return nil
}
