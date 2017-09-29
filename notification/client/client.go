package client

import (
	"github.com/tidepool-org/platform/platform"
)

type Client struct {
	client *platform.Client
}

func New(cfg *platform.Config) (*Client, error) {
	clnt, err := platform.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: clnt,
	}, nil
}
