package client

import (
	"context"

	"github.com/tidepool-org/platform/work"
)

type Client struct {
}

func New() (*Client, error) {
	return &Client{}, nil
}

func (c *Client) Create(ctx context.Context, create *work.Create) (*work.Work, error) {
	// TODO: Implement
	return nil, nil
}

func (c *Client) Get(ctx context.Context, id string) (*work.Work, error) {
	// TODO: Implement
	return nil, nil
}

func (c *Client) Process(ctx context.Context, process *work.Process) (*work.Work, error) {
	// TODO: Implement
	return nil, nil
}

func (c *Client) Repeat(ctx context.Context, id string, update *work.Repeat) (*work.Work, error) {
	// TODO: Implement
	return nil, nil
}

func (c *Client) Delete(ctx context.Context, id string) error {
	// TODO: Implement
	return nil
}
