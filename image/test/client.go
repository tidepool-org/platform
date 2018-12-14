package test

import (
	"context"

	"github.com/tidepool-org/platform/image"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
)

type ListInput struct {
	UserID     string
	Filter     *image.Filter
	Pagination *page.Pagination
}

type ListOutput struct {
	Images image.Images
	Error  error
}

type CreateInput struct {
	UserID        string
	Metadata      *image.Metadata
	ContentIntent string
	Content       *image.Content
}

type CreateOutput struct {
	Image *image.Image
	Error error
}

type CreateWithMetadataInput struct {
	UserID   string
	Metadata *image.Metadata
}

type CreateWithMetadataOutput struct {
	Image *image.Image
	Error error
}

type CreateWithContentInput struct {
	UserID        string
	ContentIntent string
	Content       *image.Content
}

type CreateWithContentOutput struct {
	Image *image.Image
	Error error
}

type GetOutput struct {
	Image *image.Image
	Error error
}

type GetMetadataOutput struct {
	Metadata *image.Metadata
	Error    error
}

type GetContentInput struct {
	ID            string
	ContentIntent *string
}

type GetContentOutput struct {
	Content *image.Content
	Error   error
}

type GetRenditionContentInput struct {
	ID        string
	Rendition *image.Rendition
}

type GetRenditionContentOutput struct {
	Content *image.Content
	Error   error
}

type PutMetadataInput struct {
	ID        string
	Condition *request.Condition
	Metadata  *image.Metadata
}

type PutMetadataOutput struct {
	Image *image.Image
	Error error
}

type PutContentInput struct {
	ID            string
	Condition     *request.Condition
	ContentIntent string
	Content       *image.Content
}

type PutContentOutput struct {
	Image *image.Image
	Error error
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
	ListInvocations                int
	ListInputs                     []ListInput
	ListStub                       func(ctx context.Context, userID string, filter *image.Filter, pagination *page.Pagination) (image.Images, error)
	ListOutputs                    []ListOutput
	ListOutput                     *ListOutput
	CreateInvocations              int
	CreateInputs                   []CreateInput
	CreateStub                     func(ctx context.Context, userID string, metadata *image.Metadata, contentIntent string, content *image.Content) (*image.Image, error)
	CreateOutputs                  []CreateOutput
	CreateOutput                   *CreateOutput
	CreateWithMetadataInvocations  int
	CreateWithMetadataInputs       []CreateWithMetadataInput
	CreateWithMetadataStub         func(ctx context.Context, userID string, metadata *image.Metadata) (*image.Image, error)
	CreateWithMetadataOutputs      []CreateWithMetadataOutput
	CreateWithMetadataOutput       *CreateWithMetadataOutput
	CreateWithContentInvocations   int
	CreateWithContentInputs        []CreateWithContentInput
	CreateWithContentStub          func(ctx context.Context, userID string, contentIntent string, content *image.Content) (*image.Image, error)
	CreateWithContentOutputs       []CreateWithContentOutput
	CreateWithContentOutput        *CreateWithContentOutput
	DeleteAllInvocations           int
	DeleteAllInputs                []string
	DeleteAllStub                  func(ctx context.Context, id string) error
	DeleteAllOutputs               []error
	DeleteAllOutput                *error
	GetInvocations                 int
	GetInputs                      []string
	GetStub                        func(ctx context.Context, id string) (*image.Image, error)
	GetOutputs                     []GetOutput
	GetOutput                      *GetOutput
	GetMetadataInvocations         int
	GetMetadataInputs              []string
	GetMetadataStub                func(ctx context.Context, id string) (*image.Metadata, error)
	GetMetadataOutputs             []GetMetadataOutput
	GetMetadataOutput              *GetMetadataOutput
	GetContentInvocations          int
	GetContentInputs               []GetContentInput
	GetContentStub                 func(ctx context.Context, id string, contentIntent *string) (*image.Content, error)
	GetContentOutputs              []GetContentOutput
	GetContentOutput               *GetContentOutput
	GetRenditionContentInvocations int
	GetRenditionContentInputs      []GetRenditionContentInput
	GetRenditionContentStub        func(ctx context.Context, id string, rendition *image.Rendition) (*image.Content, error)
	GetRenditionContentOutputs     []GetRenditionContentOutput
	GetRenditionContentOutput      *GetRenditionContentOutput
	PutMetadataInvocations         int
	PutMetadataInputs              []PutMetadataInput
	PutMetadataStub                func(ctx context.Context, id string, condition *request.Condition, metadata *image.Metadata) (*image.Image, error)
	PutMetadataOutputs             []PutMetadataOutput
	PutMetadataOutput              *PutMetadataOutput
	PutContentInvocations          int
	PutContentInputs               []PutContentInput
	PutContentStub                 func(ctx context.Context, id string, condition *request.Condition, contentIntent string, content *image.Content) (*image.Image, error)
	PutContentOutputs              []PutContentOutput
	PutContentOutput               *PutContentOutput
	DeleteInvocations              int
	DeleteInputs                   []DeleteInput
	DeleteStub                     func(ctx context.Context, id string, condition *request.Condition) (bool, error)
	DeleteOutputs                  []DeleteOutput
	DeleteOutput                   *DeleteOutput
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) List(ctx context.Context, userID string, filter *image.Filter, pagination *page.Pagination) (image.Images, error) {
	c.ListInvocations++
	c.ListInputs = append(c.ListInputs, ListInput{UserID: userID, Filter: filter, Pagination: pagination})
	if c.ListStub != nil {
		return c.ListStub(ctx, userID, filter, pagination)
	}
	if len(c.ListOutputs) > 0 {
		output := c.ListOutputs[0]
		c.ListOutputs = c.ListOutputs[1:]
		return output.Images, output.Error
	}
	if c.ListOutput != nil {
		return c.ListOutput.Images, c.ListOutput.Error
	}
	panic("List has no output")
}

func (c *Client) Create(ctx context.Context, userID string, metadata *image.Metadata, contentIntent string, content *image.Content) (*image.Image, error) {
	c.CreateInvocations++
	c.CreateInputs = append(c.CreateInputs, CreateInput{UserID: userID, Metadata: metadata, ContentIntent: contentIntent, Content: content})
	if c.CreateStub != nil {
		return c.CreateStub(ctx, userID, metadata, contentIntent, content)
	}
	if len(c.CreateOutputs) > 0 {
		output := c.CreateOutputs[0]
		c.CreateOutputs = c.CreateOutputs[1:]
		return output.Image, output.Error
	}
	if c.CreateOutput != nil {
		return c.CreateOutput.Image, c.CreateOutput.Error
	}
	panic("Create has no output")
}

func (c *Client) CreateWithMetadata(ctx context.Context, userID string, metadata *image.Metadata) (*image.Image, error) {
	c.CreateWithMetadataInvocations++
	c.CreateWithMetadataInputs = append(c.CreateWithMetadataInputs, CreateWithMetadataInput{UserID: userID, Metadata: metadata})
	if c.CreateWithMetadataStub != nil {
		return c.CreateWithMetadataStub(ctx, userID, metadata)
	}
	if len(c.CreateWithMetadataOutputs) > 0 {
		output := c.CreateWithMetadataOutputs[0]
		c.CreateWithMetadataOutputs = c.CreateWithMetadataOutputs[1:]
		return output.Image, output.Error
	}
	if c.CreateWithMetadataOutput != nil {
		return c.CreateWithMetadataOutput.Image, c.CreateWithMetadataOutput.Error
	}
	panic("CreateWithMetadata has no output")
}

func (c *Client) CreateWithContent(ctx context.Context, userID string, contentIntent string, content *image.Content) (*image.Image, error) {
	c.CreateWithContentInvocations++
	c.CreateWithContentInputs = append(c.CreateWithContentInputs, CreateWithContentInput{UserID: userID, ContentIntent: contentIntent, Content: content})
	if c.CreateWithContentStub != nil {
		return c.CreateWithContentStub(ctx, userID, contentIntent, content)
	}
	if len(c.CreateWithContentOutputs) > 0 {
		output := c.CreateWithContentOutputs[0]
		c.CreateWithContentOutputs = c.CreateWithContentOutputs[1:]
		return output.Image, output.Error
	}
	if c.CreateWithContentOutput != nil {
		return c.CreateWithContentOutput.Image, c.CreateWithContentOutput.Error
	}
	panic("CreateWithContent has no output")
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

func (c *Client) Get(ctx context.Context, id string) (*image.Image, error) {
	c.GetInvocations++
	c.GetInputs = append(c.GetInputs, id)
	if c.GetStub != nil {
		return c.GetStub(ctx, id)
	}
	if len(c.GetOutputs) > 0 {
		output := c.GetOutputs[0]
		c.GetOutputs = c.GetOutputs[1:]
		return output.Image, output.Error
	}
	if c.GetOutput != nil {
		return c.GetOutput.Image, c.GetOutput.Error
	}
	panic("Get has no output")
}

func (c *Client) GetMetadata(ctx context.Context, id string) (*image.Metadata, error) {
	c.GetMetadataInvocations++
	c.GetMetadataInputs = append(c.GetMetadataInputs, id)
	if c.GetMetadataStub != nil {
		return c.GetMetadataStub(ctx, id)
	}
	if len(c.GetMetadataOutputs) > 0 {
		output := c.GetMetadataOutputs[0]
		c.GetMetadataOutputs = c.GetMetadataOutputs[1:]
		return output.Metadata, output.Error
	}
	if c.GetMetadataOutput != nil {
		return c.GetMetadataOutput.Metadata, c.GetMetadataOutput.Error
	}
	panic("GetMetadata has no output")
}

func (c *Client) GetContent(ctx context.Context, id string, contentIntent *string) (*image.Content, error) {
	c.GetContentInvocations++
	c.GetContentInputs = append(c.GetContentInputs, GetContentInput{ID: id, ContentIntent: contentIntent})
	if c.GetContentStub != nil {
		return c.GetContentStub(ctx, id, contentIntent)
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

func (c *Client) GetRenditionContent(ctx context.Context, id string, rendition *image.Rendition) (*image.Content, error) {
	c.GetRenditionContentInvocations++
	c.GetRenditionContentInputs = append(c.GetRenditionContentInputs, GetRenditionContentInput{ID: id, Rendition: rendition})
	if c.GetRenditionContentStub != nil {
		return c.GetRenditionContentStub(ctx, id, rendition)
	}
	if len(c.GetRenditionContentOutputs) > 0 {
		output := c.GetRenditionContentOutputs[0]
		c.GetRenditionContentOutputs = c.GetRenditionContentOutputs[1:]
		return output.Content, output.Error
	}
	if c.GetRenditionContentOutput != nil {
		return c.GetRenditionContentOutput.Content, c.GetRenditionContentOutput.Error
	}
	panic("GetRenditionContent has no output")
}

func (c *Client) PutMetadata(ctx context.Context, id string, condition *request.Condition, metadata *image.Metadata) (*image.Image, error) {
	c.PutMetadataInvocations++
	c.PutMetadataInputs = append(c.PutMetadataInputs, PutMetadataInput{ID: id, Condition: condition, Metadata: metadata})
	if c.PutMetadataStub != nil {
		return c.PutMetadataStub(ctx, id, condition, metadata)
	}
	if len(c.PutMetadataOutputs) > 0 {
		output := c.PutMetadataOutputs[0]
		c.PutMetadataOutputs = c.PutMetadataOutputs[1:]
		return output.Image, output.Error
	}
	if c.PutMetadataOutput != nil {
		return c.PutMetadataOutput.Image, c.PutMetadataOutput.Error
	}
	panic("PutMetadata has no output")
}

func (c *Client) PutContent(ctx context.Context, id string, condition *request.Condition, contentIntent string, content *image.Content) (*image.Image, error) {
	c.PutContentInvocations++
	c.PutContentInputs = append(c.PutContentInputs, PutContentInput{ID: id, Condition: condition, ContentIntent: contentIntent, Content: content})
	if c.PutContentStub != nil {
		return c.PutContentStub(ctx, id, condition, contentIntent, content)
	}
	if len(c.PutContentOutputs) > 0 {
		output := c.PutContentOutputs[0]
		c.PutContentOutputs = c.PutContentOutputs[1:]
		return output.Image, output.Error
	}
	if c.PutContentOutput != nil {
		return c.PutContentOutput.Image, c.PutContentOutput.Error
	}
	panic("PutContent has no output")
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
	if len(c.CreateWithMetadataOutputs) > 0 {
		panic("CreateWithMetadataOutputs is not empty")
	}
	if len(c.CreateWithContentOutputs) > 0 {
		panic("CreateWithContentOutputs is not empty")
	}
	if len(c.DeleteAllOutputs) > 0 {
		panic("DeleteAllOutputs is not empty")
	}
	if len(c.GetOutputs) > 0 {
		panic("GetOutputs is not empty")
	}
	if len(c.GetMetadataOutputs) > 0 {
		panic("GetMetadataOutputs is not empty")
	}
	if len(c.GetContentOutputs) > 0 {
		panic("GetContentOutputs is not empty")
	}
	if len(c.GetRenditionContentOutputs) > 0 {
		panic("GetRenditionContentOutputs is not empty")
	}
	if len(c.PutMetadataOutputs) > 0 {
		panic("PutMetadataOutputs is not empty")
	}
	if len(c.PutContentOutputs) > 0 {
		panic("PutContentOutputs is not empty")
	}
	if len(c.DeleteOutputs) > 0 {
		panic("DeleteOutputs is not empty")
	}
}
