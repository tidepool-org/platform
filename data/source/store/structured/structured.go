package structured

import (
	"context"

	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
)

type Store interface {
	NewDataRepository() DataRepository
}

type DataRepository interface {
	List(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error)
	Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error)
	DestroyAll(ctx context.Context, userID string) (bool, error)

	Get(ctx context.Context, id string) (*dataSource.Source, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error)
	Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error)
}
