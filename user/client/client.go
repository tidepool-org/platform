package client

import (
	"github.com/tidepool-org/platform/platform"
)

type Client struct {
	client *platform.Client
}

func New(cfg *platform.Config, authorizeAs platform.AuthorizeAs) (*Client, error) {
	clnt, err := platform.NewClient(cfg, authorizeAs)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: clnt,
	}, nil
}
