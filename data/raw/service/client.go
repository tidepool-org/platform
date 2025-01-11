package service

import (
	"context"
	"errors"
	"io"

	dataRaw "github.com/tidepool-org/platform/data/raw"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	storeStructured "github.com/tidepool-org/platform/store/structured"
)

//go:generate mockgen -source=client.go -destination=test/client.go -package test Store
type Store interface {
	List(ctx context.Context, userID string, filter *dataRaw.Filter, pagination *page.Pagination) ([]*dataRaw.Raw, error)
	Create(ctx context.Context, userID string, dataSetID string, create *dataRaw.Create, data io.Reader) (*dataRaw.Raw, error)
	Get(ctx context.Context, id string, condition *storeStructured.Condition) (*dataRaw.Raw, error)
	GetContent(ctx context.Context, id string, condition *storeStructured.Condition) (*dataRaw.Content, error)
	Delete(ctx context.Context, id string, condition *storeStructured.Condition) (*dataRaw.Raw, error)
	DeleteMultiple(ctx context.Context, ids []string) (int, error)
	DeleteAllByDataSetID(ctx context.Context, dataSetID string) (int, error)
	DeleteAllByUserID(ctx context.Context, userID string) (int, error)
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

func (c *Client) List(ctx context.Context, userID string, filter *dataRaw.Filter, pagination *page.Pagination) ([]*dataRaw.Raw, error) {
	return c.store.List(ctx, userID, filter, pagination)
}

func (c *Client) Create(ctx context.Context, userID string, dataSetID string, create *dataRaw.Create, data io.Reader) (*dataRaw.Raw, error) {
	return c.store.Create(ctx, userID, dataSetID, create, data)
}

func (c *Client) Get(ctx context.Context, id string, condition *request.Condition) (*dataRaw.Raw, error) {
	return c.store.Get(ctx, id, storeStructured.MapCondition(condition))
}

func (c *Client) GetContent(ctx context.Context, id string, condition *request.Condition) (*dataRaw.Content, error) {
	return c.store.GetContent(ctx, id, storeStructured.MapCondition(condition))
}

func (c *Client) Delete(ctx context.Context, id string, condition *request.Condition) (*dataRaw.Raw, error) {
	return c.store.Delete(ctx, id, storeStructured.MapCondition(condition))
}

func (c *Client) DeleteMultiple(ctx context.Context, ids []string) (int, error) {
	return c.store.DeleteMultiple(ctx, ids)
}

func (c *Client) DeleteAllByDataSetID(ctx context.Context, dataSetID string) (int, error) {
	return c.store.DeleteAllByDataSetID(ctx, dataSetID)
}

func (c *Client) DeleteAllByUserID(ctx context.Context, userID string) (int, error) {
	return c.store.DeleteAllByUserID(ctx, userID)
}
