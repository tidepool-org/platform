package test

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/test"
)

type GetStatusOutput struct {
	Status *task.Status
	Error  error
}

type Client struct {
	*test.Mock
	GetStatusInvocations int
	GetStatusInputs      []auth.Context
	GetStatusOutputs     []GetStatusOutput
}

func NewClient() *Client {
	return &Client{
		Mock: test.NewMock(),
	}
}

func (c *Client) GetStatus(ctx auth.Context) (*task.Status, error) {
	c.GetStatusInvocations++

	c.GetStatusInputs = append(c.GetStatusInputs, ctx)

	if len(c.GetStatusOutputs) == 0 {
		panic("Unexpected invocation of GetStatus on Client")
	}

	output := c.GetStatusOutputs[0]
	c.GetStatusOutputs = c.GetStatusOutputs[1:]
	return output.Status, output.Error
}

func (c *Client) UnusedOutputsCount() int {
	return len(c.GetStatusOutputs)
}
