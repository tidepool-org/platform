package client

import (
	"context"

	dataRaw "github.com/tidepool-org/platform/data/raw"
	"github.com/tidepool-org/platform/page"
)

type ErrorsClient struct {
}

func NewErrors() (*ErrorsClient, error) {
	return &ErrorsClient{}, nil
}

func (e *ErrorsClient) ListErrors(ctx context.Context, userID string, dataSetID string, filter *dataRaw.ErrorsFilter, pagination *page.Pagination) (dataRaw.DataSetErrorsArray, error) {
	// TODO: Implement
	return nil, nil
}

func (e *ErrorsClient) AppendErrors(ctx context.Context, userID string, dataSetID string, sourceErrors dataRaw.SourceErrors) error {
	// TODO: Implement
	return nil
}

func (e *ErrorsClient) DeleteAll(ctx context.Context, userID string) error {
	// TODO: Implement
	return nil
}
