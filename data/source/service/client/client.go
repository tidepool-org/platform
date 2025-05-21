package client

import (
	"context"

	"github.com/tidepool-org/platform/auth"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
)

type Provider interface {
	AuthClient() auth.Client
	DataSourceStructuredStore() dataSourceStoreStructured.Store
}

type Client struct {
	Provider
}

func New(provider Provider) (*Client, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}

	return &Client{
		Provider: provider,
	}, nil
}

// FUTURE: Return ErrorResourceNotFoundWithID(userID) if userID does not exist at all

func (c *Client) List(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error) {
	return c.DataSourceStructuredStore().NewDataSourcesRepository().List(ctx, userID, filter, pagination)
}

func (c *Client) Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error) {
	return c.DataSourceStructuredStore().NewDataSourcesRepository().Create(ctx, userID, create)
}

func (c *Client) DeleteAll(ctx context.Context, userID string) error {
	_, err := c.DataSourceStructuredStore().NewDataSourcesRepository().DestroyAll(ctx, userID)
	return err
}

func (c *Client) Get(ctx context.Context, id string) (*dataSource.Source, error) {
	return c.DataSourceStructuredStore().NewDataSourcesRepository().Get(ctx, id)
}

func (c *Client) Update(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
	return c.DataSourceStructuredStore().NewDataSourcesRepository().Update(ctx, id, condition, update)
}

func (c *Client) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	return c.DataSourceStructuredStore().NewDataSourcesRepository().Destroy(ctx, id, condition)
}

func (c *Client) ListAll(ctx context.Context, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error) {
	return c.DataSourceStructuredStore().NewDataSourcesRepository().ListAll(ctx, filter, pagination)
}
