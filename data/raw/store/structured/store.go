package structured

import (
	"context"
	"io"

	dataRaw "github.com/tidepool-org/platform/data/raw"
	"github.com/tidepool-org/platform/page"
	storeStructured "github.com/tidepool-org/platform/store/structured"
)

//go:generate mockgen --build_flags=--mod=mod -source=./store.go -destination=./test/store.go -package test Store
type Store interface {
	List(ctx context.Context, userID string, filter *dataRaw.Filter, pagination *page.Pagination) ([]*dataRaw.Raw, error)
	Create(ctx context.Context, userID string, dataSetID string, create *dataRaw.Create, data io.Reader) (*dataRaw.Raw, error)
	Get(ctx context.Context, id string, condition *storeStructured.Condition) (*dataRaw.Raw, error)
	GetContent(ctx context.Context, id string, condition *storeStructured.Condition) (*dataRaw.Content, error)
	Delete(ctx context.Context, id string, condition *storeStructured.Condition) (*dataRaw.Raw, error)
	DeleteMultiple(ctx context.Context, ids []string) (int, error)
	DeleteAllByDataSetID(ctx context.Context, dataSetID string) (int, error)
	DeleteAllByUserID(ctx context.Context, userID string) (int, error)
}
