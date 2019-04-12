package test

import (
	"context"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/user"
)

type GetOutput struct {
	User  *user.User
	Error error
}

type DeleteInput struct {
	ID        string
	Delete    *user.Delete
	Condition *request.Condition
}

type DeleteOutput struct {
	Deleted bool
	Error   error
}

type Client struct {
	GetInvocations    int
	GetInputs         []string
	GetStub           func(ctx context.Context, id string) (*user.User, error)
	GetOutputs        []GetOutput
	GetOutput         *GetOutput
	DeleteInvocations int
	DeleteInputs      []DeleteInput
	DeleteStub        func(ctx context.Context, id string, deleet *user.Delete, condition *request.Condition) (bool, error)
	DeleteOutputs     []DeleteOutput
	DeleteOutput      *DeleteOutput
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

func (c *Client) Delete(ctx context.Context, id string, deleet *user.Delete, condition *request.Condition) (bool, error) {
	c.DeleteInvocations++
	c.DeleteInputs = append(c.DeleteInputs, DeleteInput{ID: id, Delete: deleet, Condition: condition})
	if c.DeleteStub != nil {
		return c.DeleteStub(ctx, id, deleet, condition)
	}
	if len(c.DeleteOutputs) > 0 {
		output := c.DeleteOutputs[0]
		c.DeleteOutputs = c.DeleteOutputs[1:]
		return output.Deleted, output.Error
	}
	if c.DeleteOutput != nil {
		return c.DeleteOutput.Deleted, c.DeleteOutput.Error
	}
	panic("Delete has no output")
}

func (c *Client) AssertOutputsEmpty() {
	if len(c.GetOutputs) > 0 {
		panic("GetOutputs is not empty")
	}
	if len(c.DeleteOutputs) > 0 {
		panic("DeleteOutputs is not empty")
	}
}
