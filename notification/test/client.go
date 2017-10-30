package test

import (
	"github.com/tidepool-org/platform/test"
)

type Client struct {
	*test.Mock
}

func NewClient() *Client {
	return &Client{
		Mock: test.NewMock(),
	}
}

func (c *Client) UnusedOutputsCount() int {
	return 0
}
