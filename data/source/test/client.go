package test

import (
	"context"

	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/page"
)

type ListInput struct {
	Context    context.Context
	UserID     string
	Filter     *dataSource.Filter
	Pagination *page.Pagination
}

type ListOutput struct {
	Sources dataSource.Sources
	Error   error
}

type CreateInput struct {
	Context context.Context
	UserID  string
	Create  *dataSource.Create
}

type CreateOutput struct {
	Source *dataSource.Source
	Error  error
}

type GetInput struct {
	Context context.Context
	ID      string
}

type GetOutput struct {
	Source *dataSource.Source
	Error  error
}

type UpdateInput struct {
	Context context.Context
	UserID  string
	Update  *dataSource.Update
}

type UpdateOutput struct {
	Source *dataSource.Source
	Error  error
}

type DeleteInput struct {
	Context context.Context
	ID      string
}

type DeleteOutput struct {
	Deleted bool
	Error   error
}

type Client struct {
	ListInvocations   int
	ListInputs        []ListInput
	ListStub          func(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.Sources, error)
	ListOutputs       []ListOutput
	ListOutput        *ListOutput
	CreateInvocations int
	CreateInputs      []CreateInput
	CreateStub        func(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error)
	CreateOutputs     []CreateOutput
	CreateOutput      *CreateOutput
	GetInvocations    int
	GetInputs         []GetInput
	GetStub           func(ctx context.Context, id string) (*dataSource.Source, error)
	GetOutputs        []GetOutput
	GetOutput         *GetOutput
	UpdateInvocations int
	UpdateInputs      []UpdateInput
	UpdateStub        func(ctx context.Context, userID string, create *dataSource.Update) (*dataSource.Source, error)
	UpdateOutputs     []UpdateOutput
	UpdateOutput      *UpdateOutput
	DeleteInvocations int
	DeleteInputs      []DeleteInput
	DeleteStub        func(ctx context.Context, id string) (bool, error)
	DeleteOutputs     []DeleteOutput
	DeleteOutput      *DeleteOutput
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) List(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.Sources, error) {
	c.ListInvocations++
	c.ListInputs = append(c.ListInputs, ListInput{Context: ctx, UserID: userID, Filter: filter, Pagination: pagination})
	if c.ListStub != nil {
		return c.ListStub(ctx, userID, filter, pagination)
	}
	if len(c.ListOutputs) > 0 {
		output := c.ListOutputs[0]
		c.ListOutputs = c.ListOutputs[1:]
		return output.Sources, output.Error
	}
	if c.ListOutput != nil {
		return c.ListOutput.Sources, c.ListOutput.Error
	}
	panic("List has no output")
}

func (c *Client) Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error) {
	c.CreateInvocations++
	c.CreateInputs = append(c.CreateInputs, CreateInput{Context: ctx, UserID: userID, Create: create})
	if c.CreateStub != nil {
		return c.CreateStub(ctx, userID, create)
	}
	if len(c.CreateOutputs) > 0 {
		output := c.CreateOutputs[0]
		c.CreateOutputs = c.CreateOutputs[1:]
		return output.Source, output.Error
	}
	if c.CreateOutput != nil {
		return c.CreateOutput.Source, c.CreateOutput.Error
	}
	panic("Create has no output")
}

func (c *Client) Get(ctx context.Context, id string) (*dataSource.Source, error) {
	c.GetInvocations++
	c.GetInputs = append(c.GetInputs, GetInput{Context: ctx, ID: id})
	if c.GetStub != nil {
		return c.GetStub(ctx, id)
	}
	if len(c.GetOutputs) > 0 {
		output := c.GetOutputs[0]
		c.GetOutputs = c.GetOutputs[1:]
		return output.Source, output.Error
	}
	if c.GetOutput != nil {
		return c.GetOutput.Source, c.GetOutput.Error
	}
	panic("Get has no output")
}

func (c *Client) Update(ctx context.Context, userID string, update *dataSource.Update) (*dataSource.Source, error) {
	c.UpdateInvocations++
	c.UpdateInputs = append(c.UpdateInputs, UpdateInput{Context: ctx, UserID: userID, Update: update})
	if c.UpdateStub != nil {
		return c.UpdateStub(ctx, userID, update)
	}
	if len(c.UpdateOutputs) > 0 {
		output := c.UpdateOutputs[0]
		c.UpdateOutputs = c.UpdateOutputs[1:]
		return output.Source, output.Error
	}
	if c.UpdateOutput != nil {
		return c.UpdateOutput.Source, c.UpdateOutput.Error
	}
	panic("Update has no output")
}

func (c *Client) Delete(ctx context.Context, id string) (bool, error) {
	c.DeleteInvocations++
	c.DeleteInputs = append(c.DeleteInputs, DeleteInput{Context: ctx, ID: id})
	if c.DeleteStub != nil {
		return c.DeleteStub(ctx, id)
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
	if len(c.ListOutputs) > 0 {
		panic("ListOutputs is not empty")
	}
	if len(c.CreateOutputs) > 0 {
		panic("CreateOutputs is not empty")
	}
	if len(c.GetOutputs) > 0 {
		panic("GetOutputs is not empty")
	}
	if len(c.UpdateOutputs) > 0 {
		panic("UpdateOutputs is not empty")
	}
	if len(c.DeleteOutputs) > 0 {
		panic("DeleteOutputs is not empty")
	}
}
