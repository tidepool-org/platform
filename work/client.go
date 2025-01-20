package work

import (
	"context"

	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
)

//go:generate mockgen -source=client.go -destination=test/client.go -package test Client
type Client interface {
	Poll(ctx context.Context, poll *Poll) ([]*Work, error)
	List(ctx context.Context, filter *Filter, pagination *page.Pagination) ([]*Work, error)
	Create(ctx context.Context, create *Create) (*Work, error)
	Get(ctx context.Context, id string, condition *request.Condition) (*Work, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *Update) (*Work, error)
	Delete(ctx context.Context, id string, condition *request.Condition) (*Work, error)
	DeleteAllByGroupID(ctx context.Context, groupID string) (int, error)
}
