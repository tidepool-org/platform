package postprocess

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/summary"
	summaryTypes "github.com/tidepool-org/platform/summary/types"
)

//go:generate mockgen -source=summarizers.go -destination=test/summarizers_mocks.go -package=test -typed

type Summarizers interface {
	UpdateSummaries(ctx context.Context, userID string) (SummariesUpdate, error)
	// RecalculateSummaries recalculates the CGM and BGM summaries of the user from all of
	// their data, rather than incrementally. The continuous summary doesn't classify
	// readings by glucose range, so it's only updated.
	RecalculateSummaries(ctx context.Context, userID string) (SummariesUpdate, error)
}

type SummariesUpdate struct {
	CGM *summaryTypes.CGMSummary
	BGM *summaryTypes.BGMSummary

	// UpdatedTypes reports each summary type the update created or recalculated
	UpdatedTypes []string

	// Deleted reports the id of each summary the update deleted, by summary type
	Deleted map[string]string
}

type summarizers struct {
	registry *summary.SummarizerRegistry
}

func NewSummarizers(registry *summary.SummarizerRegistry) (Summarizers, error) {
	if registry == nil {
		return nil, errors.New("summarizer registry is missing")
	}
	return &summarizers{registry: registry}, nil
}

func (s *summarizers) UpdateSummaries(ctx context.Context, userID string) (SummariesUpdate, error) {
	return s.updateSummaries(ctx, userID, false)
}

func (s *summarizers) RecalculateSummaries(ctx context.Context, userID string) (SummariesUpdate, error) {
	return s.updateSummaries(ctx, userID, true)
}

// updateSummaries returns the update made so far even when it reports an error, so that a change
// calculated before the error is recorded rather than lost to the retry.
func (s *summarizers) updateSummaries(ctx context.Context, userID string, recalculate bool) (SummariesUpdate, error) {
	update := SummariesUpdate{}
	var err error
	if update.CGM, err = updateSummary(ctx, summary.GetSummarizer[*summaryTypes.CGMPeriods, *summaryTypes.GlucoseBucket](s.registry), userID, recalculate, &update); err != nil {
		return update, err
	}
	if update.BGM, err = updateSummary(ctx, summary.GetSummarizer[*summaryTypes.BGMPeriods, *summaryTypes.GlucoseBucket](s.registry), userID, recalculate, &update); err != nil {
		return update, err
	}
	if _, err = summary.GetSummarizer[*summaryTypes.ContinuousPeriods, *summaryTypes.ContinuousBucket](s.registry).UpdateSummary(ctx, userID); err != nil {
		return update, errors.Wrapf(err, "unable to update %s summary", summaryTypes.SummaryTypeContinuous)
	}
	return update, nil
}

// updateSummary calculates the summary of the user and records the change made in the given update
func updateSummary[PP summaryTypes.PeriodsPt[P, PB, B], PB summaryTypes.BucketDataPt[B], P summaryTypes.Periods, B summaryTypes.BucketData](ctx context.Context, summarizer summary.Summarizer[PP, PB, P, B], userID string, recalculate bool, update *SummariesUpdate) (*summaryTypes.Summary[PP, PB, P, B], error) {
	summaryType := summaryTypes.GetType[PP, PB]()

	before, err := summarizer.GetSummary(ctx, userID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get %s summary", summaryType)
	}
	var after *summaryTypes.Summary[PP, PB, P, B]
	if recalculate {
		after, err = summarizer.RecalculateSummary(ctx, userID)
	} else {
		after, err = summarizer.UpdateSummary(ctx, userID)
	}
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

func (u *SummariesUpdate) recordUpdated(summaryType string) {
	u.UpdatedTypes = append(u.UpdatedTypes, summaryType)
}

func (u *SummariesUpdate) recordDeleted(summaryType string, summaryID string) {
	if u.Deleted == nil {
		u.Deleted = map[string]string{}
	}
	u.Deleted[summaryType] = summaryID
}
