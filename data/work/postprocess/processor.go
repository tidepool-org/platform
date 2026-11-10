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
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

type Processor struct {
	*workBase.Processor[Metadata]
	Summarizers
	ClinicsClient

	pendingBuilder  *deferredPendingBuilder
	summariesUpdate SummariesUpdate

	// patientFound records whether the clinic service found a patient record of the user
	// when the standard summaries were reported. Range-set summaries are computed only
	// then, as most users are a patient of no clinic.
	patientFound bool
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

	return &Processor{
		Processor:      processor,
		Summarizers:    dependencies.Summarizers,
		ClinicsClient:  dependencies.ClinicsClient,
		pendingBuilder: pendingBuilder,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		func() *work.ProcessResult { return p.validateWork(wrk) },
		p.absorbPending,
		p.updateSummaries,
		p.updateClinicSummaries,
		p.updateRangeSetSummaries,
		p.triggerEHRSync,
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
		GroupID: pointer.FromString(IDFromUserID(*p.Metadata().UserID)),
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
	updateSummaries := p.UpdateSummaries
	if RecalculatesSummaries(p.Metadata().Reasons) {
		updateSummaries = p.RecalculateSummaries
	}
	p.summariesUpdate, err = updateSummaries(p.Context(), *p.Metadata().UserID)

	// The changes made are recorded in the metadata before they are synced to the clinic service,
	// so that a failure between the two retries the update. They are recorded even when the
	// calculation fails partway (a partial update is returned alongside the error), riding along
	// on the failing update so the retry still reports them
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
			"deleted": p.summariesUpdate.DeletedTypes,
		}).Info("updated user summaries")
	}

	return nil
}

// updateClinicSummaries reports the standard summaries to the clinic service: those that
// changed, or every one the user has when a reason requires a sync. The report is tagged
// with the standard range set's key, so that it reaches only the patient records using it;
// updateRangeSetSummaries reports to the others. A deletion reaches every record, whatever
// its ranges.
func (p *Processor) updateClinicSummaries() *work.ProcessResult {
	workMetadata := p.Metadata()
	userID := *workMetadata.UserID
	sync := SyncsClinicSummaries(workMetadata.Reasons)
	if len(workMetadata.PendingSummaryUpdates) == 0 &&
		len(workMetadata.PendingSummaryDeletes) == 0 && !sync {

		return nil
	}

	for _, pendingDelete := range workMetadata.PendingSummaryDeletes {
		var err error
		if IsSummaryID(pendingDelete) {
			err = p.ClinicsClient.DeletePatientSummary(p.Context(), pendingDelete)
		} else {
			err = p.ClinicsClient.DeletePatientSummaryOfType(p.Context(), userID, pendingDelete)
		}
		if err != nil {
			return p.Failing(errors.Wrap(err, "unable to delete patient summary"))
		}
	}

	var cgm *summaryTypes.CGMSummary
	var bgm *summaryTypes.BGMSummary
	updates := workMetadata.PendingSummaryUpdates
	if sync || slices.Contains(updates, summaryTypes.SummaryTypeCGM) {
		cgm = p.summariesUpdate.CGM
	}
	if sync || slices.Contains(updates, summaryTypes.SummaryTypeBGM) {
		bgm = p.summariesUpdate.BGM
	}
	if cgm != nil || bgm != nil {
		key := pointer.FromString(summaryTypes.StandardRangeSetKey)
		patientSummary := clinics.NewPatientSummary(cgm, bgm)
		found, err := p.ClinicsClient.UpdatePatientSummary(p.Context(), userID, key, patientSummary)
		if err != nil {
			return p.Failing(errors.Wrap(err, "unable to update patient summary"))
		}
		p.patientFound = found
	}

	log.LoggerFromContext(p.Context()).WithFields(log.Fields{
		"updated": workMetadata.PendingSummaryUpdates,
		"deleted": workMetadata.PendingSummaryDeletes,
		"sync":    sync,
	}).Debug("synced user summaries with the clinic service")

	return nil
}

// updateRangeSetSummaries reports, for each glycemic range set other than the standard
// one that the user's patient records use, a summary computed for it, tagged with its key.
// The range sets are found afresh on each attempt from the records the clinic service
// lists, so the metadata records nothing; the clinic service accepts a report resent by a
// retry as a redelivery.
//
// A record whose ranges describe no range set platform knows is skipped and logged as an
// error, as a retry wouldn't change that; the clinic service rejects such ranges, so none
// are expected. A range set that fails to be summarized or reported doesn't stop the
// others: each is attempted, and the work then fails, so that the retry reports them all
// again.
func (p *Processor) updateRangeSetSummaries() *work.ProcessResult {
	if !p.patientFound {
		return nil
	}
	userID := *p.Metadata().UserID
	logger := log.LoggerFromContext(p.Context())

	relationships, err := p.ClinicsClient.ListClinicsForPatient(p.Context(), userID)
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to list clinics for patient"))
	}

	var keys []string
	rangeSets := map[string]summaryTypes.RangeSet{}
	for _, relationship := range relationships {
		rangeSet, err := clinics.RangeSetFor(relationship.Patient.GlycemicRanges)
		if err != nil {
			logger.WithError(err).WithField("clinicId", relationship.Clinic.Id).
				Error("unable to summarize for the glycemic ranges of a patient record")
			continue
		}
		if rangeSet.Key == summaryTypes.StandardRangeSetKey {
			continue
		}
		if _, seen := rangeSets[rangeSet.Key]; !seen {
			rangeSets[rangeSet.Key] = rangeSet
			keys = append(keys, rangeSet.Key)
		}
	}

	var errs []error
	for _, key := range keys {
		cgm, bgm, err := p.SummarizeForRangeSet(p.Context(), userID, rangeSets[key])
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "unable to summarize for range set %q", key))
			continue
		}
		if cgm == nil && bgm == nil {
			continue
		}
		patientSummary := clinics.NewPatientSummary(cgm, bgm)
		_, err = p.ClinicsClient.UpdatePatientSummary(p.Context(), userID, pointer.FromString(key), patientSummary)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "unable to update patient summary for range set %q", key))
		}
	}

	if len(keys) > 0 {
		logger.WithFields(log.Fields{
			"count":  len(keys),
			"keys":   keys,
			"failed": len(errs),
		}).Info("synced range set summaries with the clinic service")
	}
	if err := errors.Append(errs...); err != nil {
		return p.Failing(err)
	}

	return nil
}

// triggerEHRSync triggers report and flowsheet upload for patients with active subscriptions
func (p *Processor) triggerEHRSync() *work.ProcessResult {
	if !TriggersEHRSync(p.Metadata().Reasons) {
		return nil
	}

	if err := p.ClinicsClient.SyncEHRDataForPatient(p.Context(), *p.Metadata().UserID); err != nil {
		return p.Failing(errors.Wrap(err, "unable to trigger EHR sync"))
	}

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
