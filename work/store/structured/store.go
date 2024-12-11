package structured

import (
	"context"

	"github.com/tidepool-org/platform/page"
	storeStructured "github.com/tidepool-org/platform/store/structured"
	"github.com/tidepool-org/platform/work"
)

//go:generate mockgen --build_flags=--mod=mod -source=./store.go -destination=./test/store.go -package test Store
type Store interface {
	Poll(ctx context.Context, poll *work.Poll) ([]*work.Work, error)
	List(ctx context.Context, filter *work.Filter, pagination *page.Pagination) ([]*work.Work, error)
	Create(ctx context.Context, create *work.Create) (*work.Work, error)
	Get(ctx context.Context, id string, condition *storeStructured.Condition) (*work.Work, error)
	Update(ctx context.Context, id string, condition *storeStructured.Condition, update *work.Update) (*work.Work, error)
	Delete(ctx context.Context, id string, condition *storeStructured.Condition) (*work.Work, error)
	DeleteAllByGroupID(ctx context.Context, groupID string) (int, error)
}
