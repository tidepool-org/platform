package service

import (
	"context"

	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/user"
)

type ClientProvider interface {
	DataSourceStructuredStore() dataSourceStoreStructured.Store
	UserClient() user.Client
}

type Client struct {
	ClientProvider
}

func NewClient(clientProvider ClientProvider) (*Client, error) {
	if clientProvider == nil {
		return nil, errors.New("client provider is missing")
	}

	return &Client{
		ClientProvider: clientProvider,
	}, nil
}

func (c *Client) List(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.Sources, error) {
	if _, err := c.UserClient().EnsureAuthorizedUser(ctx, userID, user.OwnerPermission); err != nil {
		return nil, err
	}

	session := c.DataSourceStructuredStore().NewSession()
	defer session.Close()

	return session.List(ctx, userID, filter, pagination)
}

func (c *Client) Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error) {
	if err := c.UserClient().EnsureAuthorizedService(ctx); err != nil {
		return nil, err
	}

	session := c.DataSourceStructuredStore().NewSession()
	defer session.Close()

	return session.Create(ctx, userID, create)
}

func (c *Client) Get(ctx context.Context, id string) (*dataSource.Source, error) {
	if err := c.UserClient().EnsureAuthorized(ctx); err != nil {
		return nil, err
	}

	session := c.DataSourceStructuredStore().NewSession()
	defer session.Close()

	result, err := session.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if _, err = c.UserClient().EnsureAuthorizedUser(ctx, *result.UserID, user.OwnerPermission); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) Update(ctx context.Context, id string, update *dataSource.Update) (*dataSource.Source, error) {
	if err := c.UserClient().EnsureAuthorizedService(ctx); err != nil {
		return nil, err
	}

	session := c.DataSourceStructuredStore().NewSession()
	defer session.Close()

	return session.Update(ctx, id, update)
}

func (c *Client) Delete(ctx context.Context, id string) (bool, error) {
	if err := c.UserClient().EnsureAuthorizedService(ctx); err != nil {
		return false, err
	}

	session := c.DataSourceStructuredStore().NewSession()
	defer session.Close()

	return session.Delete(ctx, id)
}
