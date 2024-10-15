package client

import (
	"context"

	dataRaw "github.com/tidepool-org/platform/data/raw"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
)

type Client struct {
}

func New() (*Client, error) {
	return &Client{}, nil
}

func (c *Client) List(ctx context.Context, userID string, dataSetID string, pagination *page.Pagination) (dataRaw.RawArray, error) {
	// TODO: Implement
	return nil, nil
}

func (c *Client) Create(ctx context.Context, userID string, dataSetID string, content *dataRaw.Content) (*dataRaw.Raw, error) {
	// TODO: Implement
	return nil, nil
}

func (c *Client) Get(ctx context.Context, id string) (*dataRaw.Raw, error) {
	// TODO: Implement
	return nil, nil
}

func (c *Client) GetContent(ctx context.Context, id string) (*dataRaw.Content, error) {
	// TODO: Implement
	return nil, nil
}

func (c *Client) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	// TODO: Implement
	return false, nil
}

func (c *Client) DeleteAll(ctx context.Context, userID string) error {
	// TODO: Implement
	return nil
}
