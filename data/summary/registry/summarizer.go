package registry

import (
	"context"
	"github.com/tidepool-org/platform/data/summary/types"
)

type Summarizer[T types.Stats] interface {
	GetSummary(ctx context.Context, userId string) (*types.Summary[T], error)
	UpdateSummary(ctx context.Context, userId string) (*types.Summary[T], error)
}
