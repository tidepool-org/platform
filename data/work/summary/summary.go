// Package summary implements what is common to the work that calculates the summaries of a
// user and syncs them to the clinic service.
package summary

import (
	"context"
	"slices"

	clinic "github.com/tidepool-org/clinic/client"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/summary"
	summaryTypes "github.com/tidepool-org/platform/summary/types"
)

// A ClinicsClient syncs the summaries of a user to the clinic service.
type ClinicsClient interface {
	UpdatePatientSummary(ctx context.Context, patientID string,
		patientSummary *clinic.PatientSummaryV1) error
	DeletePatientSummary(ctx context.Context, summaryID string) error
}

// A SummariesUpdate is the change made by calculating the summaries of a user.
type SummariesUpdate struct {
	CGM *summaryTypes.CGMSummary
	BGM *summaryTypes.BGMSummary

	// UpdatedTypes reports each summary type the update created or recalculated
	UpdatedTypes []string

	// Deleted reports the id of each summary the update deleted, by summary type
	Deleted map[string]string
}

func (u *SummariesUpdate) recordUpdated(summaryType string) {
	u.UpdatedTypes = append(u.UpdatedTypes, summaryType)
}

func (u *SummariesUpdate) recordDeleted(summaryType string, summaryID string) {
	if u.Deleted == nil {
		u.Deleted = map[string]string{}
	}
	u.Deleted[summaryType] = summaryID
}

// CalculateSummary returns the summary of the user that calculate returns, and records the
// change calculate made to the summary in update. It is given the calculation, either
// summarizer.UpdateSummary or summarizer.RecalculateSummary, as calculate.
func CalculateSummary[
	PP summaryTypes.PeriodsPt[P, PB, B], PB summaryTypes.BucketDataPt[B],
	P summaryTypes.Periods, B summaryTypes.BucketData,
](ctx context.Context, summarizer summary.Summarizer[PP, PB, P, B],
	calculate func(ctx context.Context, userID string) (
		*summaryTypes.Summary[PP, PB, P, B], error),
	userID string, update *SummariesUpdate) (*summaryTypes.Summary[PP, PB, P, B], error) {

	summaryType := summaryTypes.GetType[PP, PB]()

	before, err := summarizer.GetSummary(ctx, userID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get %s summary", summaryType)
	}
	after, err := calculate(ctx, userID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to update %s summary", summaryType)
	}

	if after == nil {
		if before != nil {
			update.recordDeleted(summaryType, before.ID.Hex())
		}
		return nil, nil
	}
	if before == nil || !before.Dates.LastUpdatedDate.Equal(after.Dates.LastUpdatedDate) {
		update.recordUpdated(summaryType)
	}
	return after, nil
}

// RecordSummariesUpdate adds the summary types that update reports as updated to
// pendingUpdates, and the ids of the summaries it reports as deleted to pendingDeletes. It
// reports whether either was changed.
//
// These are the changes that are yet to be synced to the clinic service. Work records them
// in its metadata before syncing them, so that a failure between the two is retried.
func RecordSummariesUpdate(pendingUpdates *[]string, pendingDeletes *[]string,
	update SummariesUpdate) bool {

	changed := false
	for _, summaryType := range update.UpdatedTypes {
		if !slices.Contains(*pendingUpdates, summaryType) {
			*pendingUpdates = append(*pendingUpdates, summaryType)
			changed = true
		}
	}
	for _, summaryID := range update.Deleted {
		if !slices.Contains(*pendingDeletes, summaryID) {
			*pendingDeletes = append(*pendingDeletes, summaryID)
			changed = true
		}
	}
	if changed {
		slices.Sort(*pendingUpdates)
		slices.Sort(*pendingDeletes)
	}
	return changed
}

// SyncClinicSummaries syncs the pending changes to the summaries of the user to the clinic
// service. It deletes each summary in pendingDeletes, and updates the patient with those
// summaries of update whose type is in pendingUpdates. It reports whether there was
// anything to sync.
func SyncClinicSummaries(ctx context.Context, clinicsClient ClinicsClient, userID string,
	pendingUpdates []string, pendingDeletes []string, update SummariesUpdate) (
	bool, error) {

	if len(pendingUpdates) == 0 && len(pendingDeletes) == 0 {
		return false, nil
	}

	for _, summaryID := range pendingDeletes {
		if err := clinicsClient.DeletePatientSummary(ctx, summaryID); err != nil {
			return false, errors.Wrap(err, "unable to delete patient summary")
		}
	}

	var cgm *summaryTypes.CGMSummary
	var bgm *summaryTypes.BGMSummary
	if slices.Contains(pendingUpdates, summaryTypes.SummaryTypeCGM) {
		cgm = update.CGM
	}
	if slices.Contains(pendingUpdates, summaryTypes.SummaryTypeBGM) {
		bgm = update.BGM
	}
	if cgm != nil || bgm != nil {
		patientSummary := clinics.NewPatientSummary(cgm, bgm)
		err := clinicsClient.UpdatePatientSummary(ctx, userID, patientSummary)
		if err != nil {
			return false, errors.Wrap(err, "unable to update patient summary")
		}
	}

	return true, nil
}
