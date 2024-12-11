package client

import (
	"context"
	"errors"
	"io"

	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataRawStoreStructured "github.com/tidepool-org/platform/data/raw/store/structured"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	storeStructured "github.com/tidepool-org/platform/store/structured"
)

type ClientProvider interface {
	DataRawStructuredStore() dataRawStoreStructured.Store
}

type Client struct {
	ClientProvider
}

func NewClient(provider ClientProvider) (*Client, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}

	return &Client{
		ClientProvider: provider,
	}, nil
}

func (c *Client) List(ctx context.Context, userID string, filter *dataRaw.Filter, pagination *page.Pagination) ([]*dataRaw.Raw, error) {
	return c.DataRawStructuredStore().List(ctx, userID, filter, pagination)
}

func (c *Client) Create(ctx context.Context, userID string, dataSetID string, create *dataRaw.Create, data io.Reader) (*dataRaw.Raw, error) {
	return c.DataRawStructuredStore().Create(ctx, userID, dataSetID, create, data)
}

func (c *Client) Get(ctx context.Context, id string, condition *request.Condition) (*dataRaw.Raw, error) {
	return c.DataRawStructuredStore().Get(ctx, id, storeStructured.MapCondition(condition))
}

func (c *Client) GetContent(ctx context.Context, id string, condition *request.Condition) (*dataRaw.Content, error) {
	return c.DataRawStructuredStore().GetContent(ctx, id, storeStructured.MapCondition(condition))
}

func (c *Client) Delete(ctx context.Context, id string, condition *request.Condition) (*dataRaw.Raw, error) {
	return c.DataRawStructuredStore().Delete(ctx, id, storeStructured.MapCondition(condition))
}

func (c *Client) DeleteMultiple(ctx context.Context, ids []string) (int, error) {
	return c.DataRawStructuredStore().DeleteMultiple(ctx, ids)
}

func (c *Client) DeleteAllByDataSetID(ctx context.Context, dataSetID string) (int, error) {
	return c.DataRawStructuredStore().DeleteAllByDataSetID(ctx, dataSetID)
}

func (c *Client) DeleteAllByUserID(ctx context.Context, userID string) (int, error) {
	return c.DataRawStructuredStore().DeleteAllByUserID(ctx, userID)
}
