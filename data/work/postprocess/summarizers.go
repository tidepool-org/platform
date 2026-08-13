package postprocess

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/summary"
	summaryTypes "github.com/tidepool-org/platform/summary/types"
)

//go:generate mockgen -source=summarizers.go -destination=test/summarizers_mocks.go -package=test -typed

type Summarizers interface {
	UpdateSummaries(ctx context.Context, userID string) error
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

func (s *summarizers) UpdateSummaries(ctx context.Context, userID string) error {
	if _, err := summary.GetSummarizer[*summaryTypes.CGMPeriods, *summaryTypes.GlucoseBucket](s.registry).UpdateSummary(ctx, userID); err != nil {
		return errors.Wrapf(err, "unable to update %s summary", summaryTypes.SummaryTypeCGM)
	}
	if _, err := summary.GetSummarizer[*summaryTypes.BGMPeriods, *summaryTypes.GlucoseBucket](s.registry).UpdateSummary(ctx, userID); err != nil {
		return errors.Wrapf(err, "unable to update %s summary", summaryTypes.SummaryTypeBGM)
	}
	if _, err := summary.GetSummarizer[*summaryTypes.ContinuousPeriods, *summaryTypes.ContinuousBucket](s.registry).UpdateSummary(ctx, userID); err != nil {
		return errors.Wrapf(err, "unable to update %s summary", summaryTypes.SummaryTypeContinuous)
	}
	return nil
}
