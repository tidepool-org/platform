package postprocess

import (
	"context"
	"slices"
	"time"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	summaryTypes "github.com/tidepool-org/platform/summary/types"
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

	pendingBuilder  *deferredPendingBuilder
	summariesUpdate SummariesUpdate
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
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		func() *work.ProcessResult { return p.validateWork(wrk) },
		p.FetchUserFromWorkMetadata,
		p.absorbPending,
		p.updateSummaries,
		p.updateClinicSummaries,
		p.triggerElectronicHealthRecordSync,
	).Process(p.Delete)
}

// validateWork fails if there is a mismatch between the user id and group/serial ids
func (p *Processor) validateWork(wrk *work.Work) *work.ProcessResult {
	if err := validateIdentity(wrk.GroupID, wrk.SerialID, p.Metadata()); err != nil {
		return p.Failed(errors.Wrap(err, "work is invalid"))
	}
	return nil
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
		if err == nil && workMetadata == nil {
			err = errors.New("metadata is missing")
		}
		// The work must be scoped to the same user. This check should never fail under normal circumstances
		// given we are filtering by the group id when retrieving pending work items
		if err == nil {
			err = validateIdentity(wrk.GroupID, wrk.SerialID, workMetadata)
		}
		// The pending work item will fail when it's picked up
		if err != nil {
			log.LoggerFromContext(p.Context()).WithError(err).WithField("id", wrk.ID).Warn("work pending for the user is invalid")
			continue
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
		deleted, err := p.WorkClient().Delete(p.Context(), wrk.ID, &request.Condition{Revision: &wrk.Revision})
		if err != nil {
			return p.Failing(errors.Wrap(err, "unable to delete work"))
		}
		// The work changed since it was listed - log a warning since this shouldn't happen.
		// Work items are processed serially and there's no modification path.
		if deleted == nil {
			log.LoggerFromContext(p.Context()).WithField("id", wrk.ID).Warn("work absorbed changed before it was deleted")
		}
	}

	if len(absorbed) > 0 {
		log.LoggerFromContext(p.Context()).WithFields(log.Fields{
			"count":   len(absorbed),
			"reasons": reasons,
		}).Debug("absorbed pending work for the user")
	}

	// An upload reporting only full batches is still in progress
	if deferredUntil.After(p.Now()) && shouldDefer(reasons) {
		p.pendingBuilder.availableTime = deferredUntil
		return p.Pending()
	}

	return nil
}

func (p *Processor) updateSummaries() *work.ProcessResult {
	var err error
	p.summariesUpdate, err = p.UpdateSummaries(p.Context(), *p.User().UserID)

	// The changes made are recorded in the metadata before they are synced to the clinic service,
	// so that a failure between the two retries the update
	changed := p.Metadata().recordSummariesUpdate(p.summariesUpdate)
	if err != nil {
		return p.Failing(err)
	}
	if changed {
		if result := p.ProcessingUpdate(); result != nil {
			return result
		}
		log.LoggerFromContext(p.Context()).WithFields(log.Fields{
			"reasons": p.Metadata().Reasons,
			"updated": p.summariesUpdate.UpdatedTypes,
			"deleted": p.summariesUpdate.Deleted,
		}).Info("updated user summaries")
	}

	return nil
}

func (p *Processor) updateClinicSummaries() *work.ProcessResult {
	workMetadata := p.Metadata()
	if len(workMetadata.PendingSummaryUpdates) == 0 && len(workMetadata.PendingSummaryDeletes) == 0 {
		return nil
	}

	for _, summaryID := range workMetadata.PendingSummaryDeletes {
		if err := p.ClinicsClient.DeletePatientSummary(p.Context(), summaryID); err != nil {
			return p.Failing(errors.Wrap(err, "unable to delete patient summary"))
		}
	}
	if len(workMetadata.PendingSummaryDeletes) > 0 {
		log.LoggerFromContext(p.Context()).WithFields(log.Fields{
			"deleted": workMetadata.PendingSummaryDeletes,
		}).Debug("deleted clinic service summaries")
	}

	var cgm *summaryTypes.CGMSummary
	var bgm *summaryTypes.BGMSummary
	if slices.Contains(workMetadata.PendingSummaryUpdates, summaryTypes.SummaryTypeCGM) {
		cgm = p.summariesUpdate.CGM
	}
	if slices.Contains(workMetadata.PendingSummaryUpdates, summaryTypes.SummaryTypeBGM) {
		bgm = p.summariesUpdate.BGM
	}
	if cgm != nil || bgm != nil {
		if err := p.ClinicsClient.UpdatePatientSummary(p.Context(), *p.User().UserID, clinics.NewPatientSummary(cgm, bgm)); err != nil {
			return p.Failing(errors.Wrap(err, "unable to update patient summary"))
		}
		log.LoggerFromContext(p.Context()).WithFields(log.Fields{
			"updated": workMetadata.PendingSummaryUpdates,
		}).Debug("updated clinic service summaries")
	}

	log.LoggerFromContext(p.Context()).WithFields(log.Fields{
		"updated": workMetadata.PendingSummaryUpdates,
		"deleted": workMetadata.PendingSummaryDeletes,
	}).Info("synced user summaries with the clinic service")

	return nil
}

// triggerElectronicHealthRecordSync reports the data of the user to any electronic health record it
// is shared with. Repeating the request reports the same data again, not twice, so retries are safe.
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

// deferredPendingBuilder defers work until a time decided during processing
type deferredPendingBuilder struct {
	availableTime time.Time
}

func (d *deferredPendingBuilder) ProcessingAvailableTime(ctx context.Context, wrk *work.Work, tm time.Time) time.Time {
	if d.availableTime.After(tm) {
		return d.availableTime
	}
	return tm
}
