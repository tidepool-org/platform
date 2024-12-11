package service

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	storeStructured "github.com/tidepool-org/platform/store/structured"
	"github.com/tidepool-org/platform/work"
	workStoreStructured "github.com/tidepool-org/platform/work/store/structured"
)

type ClientProvider interface {
	WorkStructuredStore() workStoreStructured.Store
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

func (c *Client) Poll(ctx context.Context, poll *work.Poll) ([]*work.Work, error) {
	return c.WorkStructuredStore().Poll(ctx, poll)
}

func (c *Client) List(ctx context.Context, filter *work.Filter, pagination *page.Pagination) ([]*work.Work, error) {
	return c.WorkStructuredStore().List(ctx, filter, pagination)
}

func (c *Client) Create(ctx context.Context, create *work.Create) (*work.Work, error) {
	return c.WorkStructuredStore().Create(ctx, create)
}

func (c *Client) Get(ctx context.Context, id string, condition *request.Condition) (*work.Work, error) {
	return c.WorkStructuredStore().Get(ctx, id, storeStructured.MapCondition(condition))
}

func (c *Client) Update(ctx context.Context, id string, condition *request.Condition, update *work.Update) (*work.Work, error) {
	return c.WorkStructuredStore().Update(ctx, id, storeStructured.MapCondition(condition), update)
}

func (c *Client) Delete(ctx context.Context, id string, condition *request.Condition) (*work.Work, error) {
	return c.WorkStructuredStore().Delete(ctx, id, storeStructured.MapCondition(condition))
}

func (c *Client) DeleteAllByGroupID(ctx context.Context, groupID string) (int, error) {
	return c.WorkStructuredStore().DeleteAllByGroupID(ctx, groupID)
}
