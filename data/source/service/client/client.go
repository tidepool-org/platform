package client

import (
	"context"

	"github.com/tidepool-org/platform/auth"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/permission"
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

func (c *Client) List(ctx context.Context, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error) {
	if filter == nil || filter.UserID == nil {
		if err := c.AuthClient().EnsureAuthorizedService(ctx); err != nil {
			return nil, err
		}
	} else {
		if _, err := c.AuthClient().EnsureAuthorizedUser(ctx, *filter.UserID, permission.Owner); err != nil {
			return nil, err
		}
	}

	repository := c.DataSourceStructuredStore().NewDataSourcesRepository()
	return repository.List(ctx, filter, pagination)
}

func (c *Client) Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error) {
	if err := c.AuthClient().EnsureAuthorizedService(ctx); err != nil {
		return nil, err
	}

	repository := c.DataSourceStructuredStore().NewDataSourcesRepository()
	return repository.Create(ctx, userID, create)
}

func (c *Client) DeleteAll(ctx context.Context, userID string) error {
	if err := c.AuthClient().EnsureAuthorizedService(ctx); err != nil {
		return err
	}

	repository := c.DataSourceStructuredStore().NewDataSourcesRepository()
	_, err := repository.DestroyAll(ctx, userID)
	return err
}

func (c *Client) Get(ctx context.Context, id string) (*dataSource.Source, error) {
	if err := c.AuthClient().EnsureAuthorized(ctx); err != nil {
		return nil, err
	}

	repository := c.DataSourceStructuredStore().NewDataSourcesRepository()
	result, err := repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if _, err = c.AuthClient().EnsureAuthorizedUser(ctx, *result.UserID, permission.Owner); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) Update(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
	if err := c.AuthClient().EnsureAuthorizedService(ctx); err != nil {
		return nil, err
	}

	repository := c.DataSourceStructuredStore().NewDataSourcesRepository()
	return repository.Update(ctx, id, condition, update)
}

func (c *Client) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	if err := c.AuthClient().EnsureAuthorizedService(ctx); err != nil {
		return false, err
	}

	repository := c.DataSourceStructuredStore().NewDataSourcesRepository()
	return repository.Destroy(ctx, id, condition)
}
