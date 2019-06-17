package test

import (
	"context"

	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
)

type ListInput struct {
	UserID     string
	Filter     *dataSource.Filter
	Pagination *page.Pagination
}

type ListOutput struct {
	SourceArray dataSource.SourceArray
	Error       error
}

type CreateInput struct {
	UserID string
	Create *dataSource.Create
}

type CreateOutput struct {
	Source *dataSource.Source
	Error  error
}

type GetOutput struct {
	Source *dataSource.Source
	Error  error
}

type UpdateInput struct {
	ID        string
	Condition *request.Condition
	Update    *dataSource.Update
}

type UpdateOutput struct {
	Source *dataSource.Source
	Error  error
}

type DeleteInput struct {
	ID        string
	Condition *request.Condition
}

type DeleteOutput struct {
	Deleted bool
	Error   error
}

type Client struct {
	ListInvocations      int
	ListInputs           []ListInput
	ListStub             func(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error)
	ListOutputs          []ListOutput
	ListOutput           *ListOutput
	CreateInvocations    int
	CreateInputs         []CreateInput
	CreateStub           func(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error)
	CreateOutputs        []CreateOutput
	CreateOutput         *CreateOutput
	DeleteAllInvocations int
	DeleteAllInputs      []string
	DeleteAllStub        func(ctx context.Context, id string) error
	DeleteAllOutputs     []error
	DeleteAllOutput      *error
	GetInvocations       int
	GetInputs            []string
	GetStub              func(ctx context.Context, id string) (*dataSource.Source, error)
	GetOutputs           []GetOutput
	GetOutput            *GetOutput
	UpdateInvocations    int
	UpdateInputs         []UpdateInput
	UpdateStub           func(ctx context.Context, id string, condition *request.Condition, create *dataSource.Update) (*dataSource.Source, error)
	UpdateOutputs        []UpdateOutput
	UpdateOutput         *UpdateOutput
	DeleteInvocations    int
	DeleteInputs         []DeleteInput
	DeleteStub           func(ctx context.Context, id string, condition *request.Condition) (bool, error)
	DeleteOutputs        []DeleteOutput
	DeleteOutput         *DeleteOutput
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) List(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error) {
	c.ListInvocations++
	c.ListInputs = append(c.ListInputs, ListInput{UserID: userID, Filter: filter, Pagination: pagination})
	if c.ListStub != nil {
		return c.ListStub(ctx, userID, filter, pagination)
	}
	if len(c.ListOutputs) > 0 {
		output := c.ListOutputs[0]
		c.ListOutputs = c.ListOutputs[1:]
		return output.SourceArray, output.Error
	}
	if c.ListOutput != nil {
		return c.ListOutput.SourceArray, c.ListOutput.Error
	}
	panic("List has no output")
}

func (c *Client) Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error) {
	c.CreateInvocations++
	c.CreateInputs = append(c.CreateInputs, CreateInput{UserID: userID, Create: create})
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

func (c *Client) DeleteAll(ctx context.Context, userID string) error {
	c.DeleteAllInvocations++
	c.DeleteAllInputs = append(c.DeleteAllInputs, userID)
	if c.DeleteAllStub != nil {
		return c.DeleteAllStub(ctx, userID)
	}
	if len(c.DeleteAllOutputs) > 0 {
		output := c.DeleteAllOutputs[0]
		c.DeleteAllOutputs = c.DeleteAllOutputs[1:]
		return output
	}
	if c.DeleteAllOutput != nil {
		return *c.DeleteAllOutput
	}
	panic("DeleteAll has no output")
}

func (c *Client) Get(ctx context.Context, id string) (*dataSource.Source, error) {
	c.GetInvocations++
	c.GetInputs = append(c.GetInputs, id)
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

func (c *Client) Update(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
	c.UpdateInvocations++
	c.UpdateInputs = append(c.UpdateInputs, UpdateInput{ID: id, Condition: condition, Update: update})
	if c.UpdateStub != nil {
		return c.UpdateStub(ctx, id, condition, update)
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

func (c *Client) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	c.DeleteInvocations++
	c.DeleteInputs = append(c.DeleteInputs, DeleteInput{ID: id, Condition: condition})
	if c.DeleteStub != nil {
		return c.DeleteStub(ctx, id, condition)
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
	if len(c.DeleteAllOutputs) > 0 {
		panic("DeleteAllOutputs is not empty")
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
