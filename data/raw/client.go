package raw

import (
	"context"

	"github.com/tidepool-org/platform/page"
)

//go:generate mockgen --build_flags=--mod=mod -source=./client.go -destination=./test/client.go -package test Client
type Client interface {
	List(ctx context.Context, userID string, dataSetID string, pagination *page.Pagination) (RawArray, error)
	Create(ctx context.Context, userID string, dataSetID string, content *Content) (*Raw, error)
	// Get(ctx context.Context, id string) (*Raw, error)
	GetContent(ctx context.Context, id string) (*Content, error)
	// Delete(ctx context.Context, id string, condition *request.Condition) (bool, error)
	// DeleteAll(ctx context.Context, userID string) error
}
