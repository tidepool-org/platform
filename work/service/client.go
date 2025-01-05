package service

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	storeStructured "github.com/tidepool-org/platform/store/structured"
	"github.com/tidepool-org/platform/work"
)

//go:generate mockgen -source=client.go -destination=test/client.go -package test Store
type Store interface {
	Poll(ctx context.Context, poll *work.Poll) ([]*work.Work, error)
	List(ctx context.Context, filter *work.Filter, pagination *page.Pagination) ([]*work.Work, error)
	Create(ctx context.Context, create *work.Create) (*work.Work, error)
	Get(ctx context.Context, id string, condition *storeStructured.Condition) (*work.Work, error)
	Update(ctx context.Context, id string, condition *storeStructured.Condition, update *work.Update) (*work.Work, error)
	Delete(ctx context.Context, id string, condition *storeStructured.Condition) (*work.Work, error)
	DeleteAllByGroupID(ctx context.Context, groupID string) (int, error)
}

type Client struct {
	store Store
}

func NewClient(store Store) (*Client, error) {
	if store == nil {
		return nil, errors.New("store is missing")
	}
	return &Client{
		store: store,
	}, nil
}

func (c *Client) Poll(ctx context.Context, poll *work.Poll) ([]*work.Work, error) {
	return c.store.Poll(ctx, poll)
}

func (c *Client) List(ctx context.Context, filter *work.Filter, pagination *page.Pagination) ([]*work.Work, error) {
	return c.store.List(ctx, filter, pagination)
}

func (c *Client) Create(ctx context.Context, create *work.Create) (*work.Work, error) {
	return c.store.Create(ctx, create)
}

func (c *Client) Get(ctx context.Context, id string, condition *request.Condition) (*work.Work, error) {
	return c.store.Get(ctx, id, storeStructured.MapCondition(condition))
}

func (c *Client) Update(ctx context.Context, id string, condition *request.Condition, update *work.Update) (*work.Work, error) {
	return c.store.Update(ctx, id, storeStructured.MapCondition(condition), update)
}

func (c *Client) Delete(ctx context.Context, id string, condition *request.Condition) (*work.Work, error) {
	return c.store.Delete(ctx, id, storeStructured.MapCondition(condition))
}

func (c *Client) DeleteAllByGroupID(ctx context.Context, groupID string) (int, error) {
	return c.store.DeleteAllByGroupID(ctx, groupID)
}
