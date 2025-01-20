package raw

import (
	"context"
	"io"

	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
)

//go:generate mockgen -source=client.go -destination=test/client.go -package test Client
type Client interface {
	List(ctx context.Context, userID string, filter *Filter, pagination *page.Pagination) ([]*Raw, error)
	Create(ctx context.Context, userID string, dataSetID string, create *Create, data io.Reader) (*Raw, error)
	Get(ctx context.Context, id string, condition *request.Condition) (*Raw, error)
	GetContent(ctx context.Context, id string, condition *request.Condition) (*Content, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *Update) (*Raw, error)
	Delete(ctx context.Context, id string, condition *request.Condition) (*Raw, error)
	DeleteMultiple(ctx context.Context, ids []string) (int, error)
	DeleteAllByDataSetID(ctx context.Context, dataSetID string) (int, error)
	DeleteAllByUserID(ctx context.Context, userID string) (int, error)
}
