package client

import (
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/oauth"
)

type ClientDependencies struct {
	Config            *client.Config
	TokenSourceSource oauth.TokenSourceSource
}

type Client struct{}

func NewClient(clientDependencies ClientDependencies) (*Client, error) {
	return &Client{}, nil
}
