package client

import (
	"context"
	"errors"

	"github.com/tidepool-org/platform/data"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	dataSummary "github.com/tidepool-org/platform/data/summary"
)

type Client struct {
	summarizerRegistry *dataSummary.SummarizerRegistry
}

func New(dataStore *dataStoreMongo.Store) (*Client, error) {
	if dataStore == nil {
		return nil, errors.New("data store is missing")
	}

	return &Client{
		summarizerRegistry: dataSummary.New(
			dataStore.NewSummaryRepository().GetStore(),
			dataStore.NewDataRepository(),
		),
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
