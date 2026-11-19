package user

import (
	"context"

	dataWorkSummary "github.com/tidepool-org/platform/data/work/summary"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/summary"
	summaryTypes "github.com/tidepool-org/platform/summary/types"
)

//go:generate mockgen -source=summarizers.go -destination=test/summarizers_mocks.go -package=test -typed

type SummariesUpdate = dataWorkSummary.SummariesUpdate

type Summarizers interface {
	// RecalculateSummaries recalculates the CGM and BGM summaries of the user from all of
	// their data, rather than incrementally. The continuous summary doesn't classify
	// readings by glucose range, so it's left as it is.
	//
	// It returns the update made so far even when it reports an error, so that a change
	// calculated before the error is recorded rather than lost to the retry.
	RecalculateSummaries(ctx context.Context, userID string) (SummariesUpdate, error)
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

func (s *summarizers) RecalculateSummaries(ctx context.Context, userID string) (
	SummariesUpdate, error) {

	update := SummariesUpdate{}
	var err error

	cgm := summary.GetSummarizer[*summaryTypes.CGMPeriods, *summaryTypes.GlucoseBucket](
		s.registry)
	update.CGM, err = dataWorkSummary.CalculateSummary(ctx, cgm, cgm.RecalculateSummary,
		userID, &update)
	if err != nil {
		return update, err
	}

	bgm := summary.GetSummarizer[*summaryTypes.BGMPeriods, *summaryTypes.GlucoseBucket](
		s.registry)
	update.BGM, err = dataWorkSummary.CalculateSummary(ctx, bgm, bgm.RecalculateSummary,
		userID, &update)
	if err != nil {
		return update, err
	}

	return update, nil
}
