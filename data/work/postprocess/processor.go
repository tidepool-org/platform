package postprocess

import (
	"context"
	"slices"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	userWork "github.com/tidepool-org/platform/user/work"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

type UserMixin = userWork.MixinFromWork

type Processor struct {
	*workBase.Processor[Metadata]
	UserMixin
	Summarizers
	ClinicsClient

	pendingBuilder *deferredPendingBuilder
	wrk            *work.Work
}

func NewProcessor(dependencies Dependencies) (*Processor, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	pendingBuilder := &deferredPendingBuilder{}
	processResultBuilder := &workBase.ProcessResultBuilder{
		ProcessResultPendingBuilder: pendingBuilder,
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
	userMixin, err := userWork.NewMixinFromWork(processor, dependencies.UserClient, &processor.Metadata().Metadata)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create user mixin")
	}

	return &Processor{
		Processor:      processor,
		UserMixin:      userMixin,
		Summarizers:    dependencies.Summarizers,
		ClinicsClient:  dependencies.ClinicsClient,
		pendingBuilder: pendingBuilder,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	p.wrk = wrk
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.FetchUserFromWorkMetadata,
		p.absorbPending,
		p.updateSummaries,
		p.triggerElectronicHealthRecordSync,
	).Process(p.Delete)
}

// absorbPending absorbs reasons of other pending work items in this serial group
//
// The reasons are persisted before the work is deleted, so that a failure between the
// two leaves the reasons reported twice rather than not at all.
func (p *Processor) absorbPending() *work.ProcessResult {
	filter := &work.Filter{
		Types:   pointer.FromAny([]string{Type}),
		State:   pointer.FromAny(work.StatePending),
		GroupID: pointer.FromString(IDFromUserID(*p.User().UserID)),
	}

	wrks, err := page.Collect(func(pagination page.Pagination) ([]*work.Work, error) {
		return p.WorkClient().List(p.Context(), filter, &pagination)
	})
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to list work"))
	}

	var absorbed []*work.Work
	var deferredUntil time.Time
	reasons := p.Metadata().Reasons
	for _, wrk := range wrks {
		workMetadata, err := metadata.Decode[Metadata](p.Context(), wrk.Metadata)
		if err != nil {
			return p.Failed(errors.Wrap(err, "unable to decode metadata"))
		} else if workMetadata == nil {
			return p.Failed(errors.New("metadata is missing"))
		}

		absorbed = append(absorbed, wrk)
		reasons = normalizeReasons(reasons, workMetadata.Reasons)
		if wrk.ProcessingAvailableTime.After(deferredUntil) {
			deferredUntil = wrk.ProcessingAvailableTime
		}
	}
	if len(absorbed) == 0 {
		return nil
	}

	p.Metadata().Reasons = reasons
	if result := p.ProcessingUpdate(); result != nil {
		return result
	}

	for _, wrk := range absorbed {
		if _, err = p.WorkClient().Delete(p.Context(), wrk.ID, nil); err != nil {
			return p.Failing(errors.Wrap(err, "unable to delete work"))
		}
	}

	if len(absorbed) > 0 {
		log.LoggerFromContext(p.Context()).WithFields(log.Fields{
			"count":   len(absorbed),
			"reasons": reasons,
		}).Debug("absorbed pending work for the user")
	}

	// An upload reporting only full batches is still in progress
	if deferredUntil.After(p.Now()) && shouldDeffer(reasons) {
		p.pendingBuilder.availableTime = deferredUntil
		return p.Pending()
	}

	return nil
}

func (p *Processor) updateSummaries() *work.ProcessResult {
	if err := p.UpdateSummaries(p.Context(), *p.User().UserID); err != nil {
		return p.Failing(err)
	}

	log.LoggerFromContext(p.Context()).WithField("reasons", p.Metadata().Reasons).Info("calculated the summaries of the user")

	return nil
}

// triggerElectronicHealthRecordSync reports the data of the user to any electronic health record it is
// shared with, after the summaries it reports are calculated. It is requested at least once per change
// reported, as a request repeated reports the same data again rather than reporting it twice.
func (p *Processor) triggerElectronicHealthRecordSync() *work.ProcessResult {
	if !TriggersEHRSync(p.Metadata().Reasons) {
		return nil
	}

	if err := p.ClinicsClient.SyncEHRDataForPatient(p.Context(), *p.User().UserID); err != nil {
		return p.Failing(errors.Wrap(err, "unable to trigger EHR sync"))
	}

	log.LoggerFromContext(p.Context()).Info("triggerred EHR sync")

	return nil
}

// deferredPendingBuilder defers work until a time decided while processing, rather than by a duration
// fixed when the processor is created
type deferredPendingBuilder struct {
	availableTime time.Time
}

func (d *deferredPendingBuilder) ProcessingAvailableTime(ctx context.Context, wrk *work.Work, tm time.Time) time.Time {
	if d.availableTime.After(tm) {
		return d.availableTime
	}
	return tm
}

// shouldDeffer returns true when jellyfish uploads a full batch
func shouldDeffer(reasons []string) bool {
	return !slices.ContainsFunc(reasons, func(reason string) bool { return reason != ReasonLegacyDataAdded })
}
