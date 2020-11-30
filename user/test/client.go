package test

import (
	"context"

	"github.com/tidepool-org/platform/user"
)

type GetOutput struct {
	User  *user.User
	Error error
}

type Client struct {
	GetInvocations    int
	GetInputs         []string
	GetStub           func(ctx context.Context, id string) (*user.User, error)
	GetOutputs        []GetOutput
	GetOutput         *GetOutput
	DeleteInvocations int
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Get(ctx context.Context, id string) (*user.User, error) {
	c.GetInvocations++
	c.GetInputs = append(c.GetInputs, id)
	if c.GetStub != nil {
		return c.GetStub(ctx, id)
	}
	if len(c.GetOutputs) > 0 {
		output := c.GetOutputs[0]
		c.GetOutputs = c.GetOutputs[1:]
		return output.User, output.Error
	}
	if c.GetOutput != nil {
		return c.GetOutput.User, c.GetOutput.Error
	}
	panic("Get has no output")
}

func (c *Client) AssertOutputsEmpty() {
	if len(c.GetOutputs) > 0 {
		panic("GetOutputs is not empty")
	}
}
