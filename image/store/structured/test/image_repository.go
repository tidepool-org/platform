package test

import (
	"context"

	"github.com/tidepool-org/platform/image"
	imageStoreStructured "github.com/tidepool-org/platform/image/store/structured"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
)

type ListInput struct {
	UserID     string
	Filter     *image.Filter
	Pagination *page.Pagination
}

type ListOutput struct {
	ImageArray image.ImageArray
	Error      error
}

type CreateInput struct {
	UserID   string
	Metadata *image.Metadata
}

type CreateOutput struct {
	Image *image.Image
	Error error
}

type DeleteAllOutput struct {
	Deleted bool
	Error   error
}

type DestroyAllOutput struct {
	Destroyed bool
	Error     error
}

type GetInput struct {
	ID        string
	Condition *request.Condition
}

type GetOutput struct {
	Image *image.Image
	Error error
}

type UpdateInput struct {
	ID        string
	Condition *request.Condition
	Update    *imageStoreStructured.Update
}

type UpdateOutput struct {
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

type DestroyInput struct {
	ID        string
	Condition *request.Condition
}

type DestroyOutput struct {
	Destroyed bool
	Error     error
}

type ImageRepository struct {
	*test.Closer
	ListInvocations       int
	ListInputs            []ListInput
	ListStub              func(ctx context.Context, userID string, filter *image.Filter, pagination *page.Pagination) (image.ImageArray, error)
	ListOutputs           []ListOutput
	ListOutput            *ListOutput
	CreateInvocations     int
	CreateInputs          []CreateInput
	CreateStub            func(ctx context.Context, userID string, metadata *image.Metadata) (*image.Image, error)
	CreateOutputs         []CreateOutput
	CreateOutput          *CreateOutput
	DeleteAllInvocations  int
	DeleteAllInputs       []string
	DeleteAllStub         func(ctx context.Context, userID string) (bool, error)
	DeleteAllOutputs      []DeleteAllOutput
	DeleteAllOutput       *DeleteAllOutput
	DestroyAllInvocations int
	DestroyAllInputs      []string
	DestroyAllStub        func(ctx context.Context, userID string) (bool, error)
	DestroyAllOutputs     []DestroyAllOutput
	DestroyAllOutput      *DestroyAllOutput
	GetInvocations        int
	GetInputs             []GetInput
	GetStub               func(ctx context.Context, id string, condition *request.Condition) (*image.Image, error)
	GetOutputs            []GetOutput
	GetOutput             *GetOutput
	UpdateInvocations     int
	UpdateInputs          []UpdateInput
	UpdateStub            func(ctx context.Context, id string, condition *request.Condition, update *imageStoreStructured.Update) (*image.Image, error)
	UpdateOutputs         []UpdateOutput
	UpdateOutput          *UpdateOutput
	DeleteInvocations     int
	DeleteInputs          []DeleteInput
	DeleteStub            func(ctx context.Context, id string, condition *request.Condition) (bool, error)
	DeleteOutputs         []DeleteOutput
	DeleteOutput          *DeleteOutput
	DestroyInvocations    int
	DestroyInputs         []DestroyInput
	DestroyStub           func(ctx context.Context, id string, condition *request.Condition) (bool, error)
	DestroyOutputs        []DestroyOutput
	DestroyOutput         *DestroyOutput
}

func NewImageRepository() *ImageRepository {
	return &ImageRepository{
		Closer: test.NewCloser(),
	}
}

func (i *ImageRepository) List(ctx context.Context, userID string, filter *image.Filter, pagination *page.Pagination) (image.ImageArray, error) {
	i.ListInvocations++
	i.ListInputs = append(i.ListInputs, ListInput{UserID: userID, Filter: filter, Pagination: pagination})
	if i.ListStub != nil {
		return i.ListStub(ctx, userID, filter, pagination)
	}
	if len(i.ListOutputs) > 0 {
		output := i.ListOutputs[0]
		i.ListOutputs = i.ListOutputs[1:]
		return output.ImageArray, output.Error
	}
	if i.ListOutput != nil {
		return i.ListOutput.ImageArray, i.ListOutput.Error
	}
	panic("List has no output")
}

func (i *ImageRepository) Create(ctx context.Context, userID string, metadata *image.Metadata) (*image.Image, error) {
	i.CreateInvocations++
	i.CreateInputs = append(i.CreateInputs, CreateInput{UserID: userID, Metadata: metadata})
	if i.CreateStub != nil {
		return i.CreateStub(ctx, userID, metadata)
	}
	if len(i.CreateOutputs) > 0 {
		output := i.CreateOutputs[0]
		i.CreateOutputs = i.CreateOutputs[1:]
		return output.Image, output.Error
	}
	if i.CreateOutput != nil {
		return i.CreateOutput.Image, i.CreateOutput.Error
	}
	panic("Create has no output")
}

func (i *ImageRepository) DeleteAll(ctx context.Context, userID string) (bool, error) {
	i.DeleteAllInvocations++
	i.DeleteAllInputs = append(i.DeleteAllInputs, userID)
	if i.DeleteAllStub != nil {
		return i.DeleteAllStub(ctx, userID)
	}
	if len(i.DeleteAllOutputs) > 0 {
		output := i.DeleteAllOutputs[0]
		i.DeleteAllOutputs = i.DeleteAllOutputs[1:]
		return output.Deleted, output.Error
	}
	if i.DeleteAllOutput != nil {
		return i.DeleteAllOutput.Deleted, i.DeleteAllOutput.Error
	}
	panic("DeleteAll has no output")
}

func (i *ImageRepository) DestroyAll(ctx context.Context, userID string) (bool, error) {
	i.DestroyAllInvocations++
	i.DestroyAllInputs = append(i.DestroyAllInputs, userID)
	if i.DestroyAllStub != nil {
		return i.DestroyAllStub(ctx, userID)
	}
	if len(i.DestroyAllOutputs) > 0 {
		output := i.DestroyAllOutputs[0]
		i.DestroyAllOutputs = i.DestroyAllOutputs[1:]
		return output.Destroyed, output.Error
	}
	if i.DestroyAllOutput != nil {
		return i.DestroyAllOutput.Destroyed, i.DestroyAllOutput.Error
	}
	panic("DestroyAll has no output")
}

func (i *ImageRepository) Get(ctx context.Context, id string, condition *request.Condition) (*image.Image, error) {
	i.GetInvocations++
	i.GetInputs = append(i.GetInputs, GetInput{ID: id, Condition: condition})
	if i.GetStub != nil {
		return i.GetStub(ctx, id, condition)
	}
	if len(i.GetOutputs) > 0 {
		output := i.GetOutputs[0]
		i.GetOutputs = i.GetOutputs[1:]
		return output.Image, output.Error
	}
	if i.GetOutput != nil {
		return i.GetOutput.Image, i.GetOutput.Error
	}
	panic("Get has no output")
}

func (i *ImageRepository) Update(ctx context.Context, id string, condition *request.Condition, update *imageStoreStructured.Update) (*image.Image, error) {
	i.UpdateInvocations++
	i.UpdateInputs = append(i.UpdateInputs, UpdateInput{ID: id, Condition: condition, Update: update})
	if i.UpdateStub != nil {
		return i.UpdateStub(ctx, id, condition, update)
	}
	if len(i.UpdateOutputs) > 0 {
		output := i.UpdateOutputs[0]
		i.UpdateOutputs = i.UpdateOutputs[1:]
		return output.Image, output.Error
	}
	if i.UpdateOutput != nil {
		return i.UpdateOutput.Image, i.UpdateOutput.Error
	}
	panic("Update has no output")
}

func (i *ImageRepository) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	i.DeleteInvocations++
	i.DeleteInputs = append(i.DeleteInputs, DeleteInput{ID: id, Condition: condition})
	if i.DeleteStub != nil {
		return i.DeleteStub(ctx, id, condition)
	}
	if len(i.DeleteOutputs) > 0 {
		output := i.DeleteOutputs[0]
		i.DeleteOutputs = i.DeleteOutputs[1:]
		return output.Deleted, output.Error
	}
	if i.DeleteOutput != nil {
		return i.DeleteOutput.Deleted, i.DeleteOutput.Error
	}
	panic("Delete has no output")
}

func (i *ImageRepository) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	i.DestroyInvocations++
	i.DestroyInputs = append(i.DestroyInputs, DestroyInput{ID: id, Condition: condition})
	if i.DestroyStub != nil {
		return i.DestroyStub(ctx, id, condition)
	}
	if len(i.DestroyOutputs) > 0 {
		output := i.DestroyOutputs[0]
		i.DestroyOutputs = i.DestroyOutputs[1:]
		return output.Destroyed, output.Error
	}
	if i.DestroyOutput != nil {
		return i.DestroyOutput.Destroyed, i.DestroyOutput.Error
	}
	panic("Destroy has no output")
}

func (i *ImageRepository) AssertOutputsEmpty() {
	i.Closer.AssertOutputsEmpty()
	if len(i.ListOutputs) > 0 {
		panic("ListOutputs is not empty")
	}
	if len(i.CreateOutputs) > 0 {
		panic("CreateOutputs is not empty")
	}
	if len(i.DeleteAllOutputs) > 0 {
		panic("DeleteAllOutputs is not empty")
	}
	if len(i.DestroyAllOutputs) > 0 {
		panic("DestroyAllOutputs is not empty")
	}
	if len(i.GetOutputs) > 0 {
		panic("GetOutputs is not empty")
	}
	if len(i.UpdateOutputs) > 0 {
		panic("UpdateOutputs is not empty")
	}
	if len(i.DeleteOutputs) > 0 {
		panic("DeleteOutputs is not empty")
	}
	if len(i.DestroyOutputs) > 0 {
		panic("DestroyOutputs is not empty")
	}
}
