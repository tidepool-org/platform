package raw

import (
	"context"

	"github.com/tidepool-org/platform/page"
)

//go:generate mockgen --build_flags=--mod=mod -source=./errors_client.go -destination=./test/errors_client.go -package test ErrorsClient
type ErrorsClient interface {
	ListErrors(ctx context.Context, userID string, dataSetID string, filter *ErrorsFilter, pagination *page.Pagination) (DataSetErrorsArray, error)
	AppendErrors(ctx context.Context, userID string, dataSetID string, sourceErrors SourceErrors) error
	DeleteAll(ctx context.Context, userID string) error
}
