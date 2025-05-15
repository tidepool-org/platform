package service

import (
	"context"

	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/page"
)

type Client interface {
	dataSource.Client

	ListAll(ctx context.Context, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error)
}
