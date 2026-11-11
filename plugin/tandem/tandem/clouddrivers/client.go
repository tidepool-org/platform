package clouddrivers

import (
	"github.com/tidepool-org/platform/platform"
)

type Client struct{}

func NewClient(cfg *platform.Config) (*Client, error) {
	return &Client{}, nil
}
