package client

import (
	"context"
	"errors"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/summary"
)

type Client struct {
	summarizerRegistry *summary.SummarizerRegistry
}

func New(summarizerRegistry *summary.SummarizerRegistry) (*Client, error) {
	if summarizerRegistry == nil {
		return nil, errors.New("summarizer registry missing")
	}

	return &Client{
		summarizerRegistry: summarizerRegistry,
	}, nil
}

func (s *Client) CheckDataUpdatesSummary(datumArray data.Data, updatesSummary map[string]struct{}) {
	for _, datum := range datumArray {
		summary.CheckDatumUpdatesSummary(updatesSummary, datum)
	}
}

func (s *Client) MaybeUpdateSummary(ctx context.Context, userID string, reason string, updatesSummary map[string]struct{}) {
	summary.MaybeUpdateSummary(ctx, s.summarizerRegistry, updatesSummary, userID, reason)
}
