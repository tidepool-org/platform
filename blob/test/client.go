package test

import (
	"context"

	"github.com/tidepool-org/platform/blob"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
)

type ListInput struct {
	UserID     string
	Filter     *blob.Filter
	Pagination *page.Pagination
}

type ListOutput struct {
	BlobArray blob.BlobArray
	Error     error
}

type CreateInput struct {
	UserID  string
	Content *blob.Content
}

type CreateOutput struct {
	Blob  *blob.Blob
	Error error
}

type CreateDeviceLogsInput struct {
	UserID  string
	Content *blob.DeviceLogsContent
}

type CreateDeviceLogsOutput struct {
	Blob  *blob.DeviceLogsBlob
	Error error
}

type GetOutput struct {
	Blob  *blob.Blob
	Error error
}

type GetContentOutput struct {
	Content *blob.Content
	Error   error
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
	ListInvocations             int
	ListInputs                  []ListInput
	ListStub                    func(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.BlobArray, error)
	ListOutputs                 []ListOutput
	ListOutput                  *ListOutput
	CreateInvocations           int
	CreateInputs                []CreateInput
	CreateDeviceLogsInvocations int
	CreateDeviceLogsInputs      []CreateDeviceLogsInput
	CreateStub                  func(ctx context.Context, userID string, content *blob.Content) (*blob.Blob, error)
	CreateOutputs               []CreateOutput
	CreateOutput                *CreateOutput
	CreatDeviceLogsStub         func(ctx context.Context, userID string, content *blob.DeviceLogsContent) (*blob.DeviceLogsBlob, error)
	CreateDeviceLogsOutputs     []CreateDeviceLogsOutput
	CreateDeviceLogsOutput      *CreateDeviceLogsOutput
	DeleteAllInvocations        int
	DeleteAllInputs             []string
	DeleteAllStub               func(ctx context.Context, id string) error
	DeleteAllOutputs            []error
	DeleteAllOutput             *error
	GetInvocations              int
	GetInputs                   []string
	GetStub                     func(ctx context.Context, id string) (*blob.Blob, error)
	GetOutputs                  []GetOutput
	GetOutput                   *GetOutput
	GetContentInvocations       int
	GetContentInputs            []string
	GetContentStub              func(ctx context.Context, id string) (*blob.Content, error)
	GetContentOutputs           []GetContentOutput
	GetContentOutput            *GetContentOutput
	DeleteInvocations           int
	DeleteInputs                []DeleteInput
	DeleteStub                  func(ctx context.Context, id string, condition *request.Condition) (bool, error)
	DeleteOutputs               []DeleteOutput
	DeleteOutput                *DeleteOutput
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) List(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.BlobArray, error) {
	c.ListInvocations++
	c.ListInputs = append(c.ListInputs, ListInput{UserID: userID, Filter: filter, Pagination: pagination})
	if c.ListStub != nil {
		return c.ListStub(ctx, userID, filter, pagination)
	}
	if len(c.ListOutputs) > 0 {
		output := c.ListOutputs[0]
		c.ListOutputs = c.ListOutputs[1:]
		return output.BlobArray, output.Error
	}
	if c.ListOutput != nil {
		return c.ListOutput.BlobArray, c.ListOutput.Error
	}
	panic("List has no output")
}

func (c *Client) Create(ctx context.Context, userID string, content *blob.Content) (*blob.Blob, error) {
	c.CreateInvocations++
	c.CreateInputs = append(c.CreateInputs, CreateInput{UserID: userID, Content: content})
	if c.CreateStub != nil {
		return c.CreateStub(ctx, userID, content)
	}
	if len(c.CreateOutputs) > 0 {
		output := c.CreateOutputs[0]
		c.CreateOutputs = c.CreateOutputs[1:]
		return output.Blob, output.Error
	}
	if c.CreateOutput != nil {
		return c.CreateOutput.Blob, c.CreateOutput.Error
	}
	panic("Create has no output")
}

func (c *Client) CreateDeviceLogs(ctx context.Context, userID string, content *blob.DeviceLogsContent) (*blob.DeviceLogsBlob, error) {
	c.CreateDeviceLogsInvocations++
	c.CreateDeviceLogsInputs = append(c.CreateDeviceLogsInputs, CreateDeviceLogsInput{UserID: userID, Content: content})
	if c.CreatDeviceLogsStub != nil {
		return c.CreatDeviceLogsStub(ctx, userID, content)
	}
	if len(c.CreateDeviceLogsOutputs) > 0 {
		output := c.CreateDeviceLogsOutputs[0]
		c.CreateDeviceLogsOutputs = c.CreateDeviceLogsOutputs[1:]
		return output.Blob, output.Error
	}
	if c.CreateDeviceLogsOutput != nil {
		return c.CreateDeviceLogsOutput.Blob, c.CreateDeviceLogsOutput.Error
	}
	panic("CreateDeviceLogs has no output")
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

func (c *Client) Get(ctx context.Context, id string) (*blob.Blob, error) {
	c.GetInvocations++
	c.GetInputs = append(c.GetInputs, id)
	if c.GetStub != nil {
		return c.GetStub(ctx, id)
	}
	if len(c.GetOutputs) > 0 {
		output := c.GetOutputs[0]
		c.GetOutputs = c.GetOutputs[1:]
		return output.Blob, output.Error
	}
	if c.GetOutput != nil {
		return c.GetOutput.Blob, c.GetOutput.Error
	}
	panic("Get has no output")
}

func (c *Client) GetContent(ctx context.Context, id string) (*blob.Content, error) {
	c.GetContentInvocations++
	c.GetContentInputs = append(c.GetContentInputs, id)
	if c.GetContentStub != nil {
		return c.GetContentStub(ctx, id)
	}
	if len(c.GetContentOutputs) > 0 {
		output := c.GetContentOutputs[0]
		c.GetContentOutputs = c.GetContentOutputs[1:]
		return output.Content, output.Error
	}
	if c.GetContentOutput != nil {
		return c.GetContentOutput.Content, c.GetContentOutput.Error
	}
	panic("GetContent has no output")
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
	if len(c.GetContentOutputs) > 0 {
		panic("GetContentOutputs is not empty")
	}
	if len(c.DeleteOutputs) > 0 {
		panic("DeleteOutputs is not empty")
	}
}
