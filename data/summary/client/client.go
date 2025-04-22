package client

import (
	"context"
	"errors"

	"github.com/tidepool-org/platform/data"
	dataSummary "github.com/tidepool-org/platform/data/summary"
)

type Client struct {
	summarizerRegistry *dataSummary.SummarizerRegistry
}

func New(summarizerRegistry *dataSummary.SummarizerRegistry) (*Client, error) {
	if summarizerRegistry == nil {
		return nil, errors.New("summarizer registry missing")
	}

	return &Client{
		summarizerRegistry: summarizerRegistry,
	}, nil
}

func (s *Client) CheckDataUpdatesSummary(datumArray data.Data, updatesSummary map[string]struct{}) {
	for _, datum := range datumArray {
		dataSummary.CheckDatumUpdatesSummary(updatesSummary, datum)
	}
}

func (s *Client) MaybeUpdateSummary(ctx context.Context, userID string, reason string, updatesSummary map[string]struct{}) {
	dataSummary.MaybeUpdateSummary(ctx, s.summarizerRegistry, updatesSummary, userID, reason)
}
