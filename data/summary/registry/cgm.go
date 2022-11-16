package registry

import (
	"context"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/summary/store"
	"github.com/tidepool-org/platform/data/summary/types"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type CGMSummarizer struct {
	deviceData dataStore.DataRepository
	summaries  *store.Repo[types.CGMStats]
}

// Compile time interface check
var _ Summarizer[types.CGMStats] = &CGMSummarizer{}

func NewCGMSummarizer(collection *storeStructuredMongo.Repository, deviceData dataStore.DataRepository) Summarizer[types.CGMStats] {
	return &CGMSummarizer{
		deviceData: deviceData,
		summaries:  store.New[types.CGMStats](collection),
	}
}

func (c *CGMSummarizer) GetSummary(ctx context.Context, userId string) (*types.Summary[types.CGMStats], error) {
	return c.summaries.GetSummary(ctx, userId)
}

func (c *CGMSummarizer) UpdateSummary(ctx context.Context, userId string) (*types.Summary[types.CGMStats], error) {
	old, err := c.GetSummary(ctx, userId)
	if err != nil {
		return nil, err
	}
	// Type specific calculations go here
	// c.deviceData.GetDataRange(ctx, userId, "cgm" ...)
	return old, nil
}
