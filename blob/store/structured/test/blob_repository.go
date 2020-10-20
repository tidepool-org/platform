package test

import (
	"context"

	"github.com/tidepool-org/platform/blob"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
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
	UserID string
	Create *blobStoreStructured.Create
}

type CreateOutput struct {
	Blob  *blob.Blob
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
	Blob  *blob.Blob
	Error error
}

type UpdateInput struct {
	ID        string
	Condition *request.Condition
	Update    *blobStoreStructured.Update
}

type UpdateOutput struct {
	Blob  *blob.Blob
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

type BlobRepository struct {
	ListInvocations       int
	ListInputs            []ListInput
	ListStub              func(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.BlobArray, error)
	ListOutputs           []ListOutput
	ListOutput            *ListOutput
	CreateInvocations     int
	CreateInputs          []CreateInput
	CreateStub            func(ctx context.Context, userID string, create *blobStoreStructured.Create) (*blob.Blob, error)
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
	GetStub               func(ctx context.Context, id string, condition *request.Condition) (*blob.Blob, error)
	GetOutputs            []GetOutput
	GetOutput             *GetOutput
	UpdateInvocations     int
	UpdateInputs          []UpdateInput
	UpdateStub            func(ctx context.Context, id string, condition *request.Condition, update *blobStoreStructured.Update) (*blob.Blob, error)
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

func NewBlobRepository() *BlobRepository {
	return &BlobRepository{}
}

func (b *BlobRepository) List(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.BlobArray, error) {
	b.ListInvocations++
	b.ListInputs = append(b.ListInputs, ListInput{UserID: userID, Filter: filter, Pagination: pagination})
	if b.ListStub != nil {
		return b.ListStub(ctx, userID, filter, pagination)
	}
	if len(b.ListOutputs) > 0 {
		output := b.ListOutputs[0]
		b.ListOutputs = b.ListOutputs[1:]
		return output.BlobArray, output.Error
	}
	if b.ListOutput != nil {
		return b.ListOutput.BlobArray, b.ListOutput.Error
	}
	panic("List has no output")
}

func (b *BlobRepository) Create(ctx context.Context, userID string, create *blobStoreStructured.Create) (*blob.Blob, error) {
	b.CreateInvocations++
	b.CreateInputs = append(b.CreateInputs, CreateInput{UserID: userID, Create: create})
	if b.CreateStub != nil {
		return b.CreateStub(ctx, userID, create)
	}
	if len(b.CreateOutputs) > 0 {
		output := b.CreateOutputs[0]
		b.CreateOutputs = b.CreateOutputs[1:]
		return output.Blob, output.Error
	}
	if b.CreateOutput != nil {
		return b.CreateOutput.Blob, b.CreateOutput.Error
	}
	panic("Create has no output")
}

func (b *BlobRepository) DeleteAll(ctx context.Context, userID string) (bool, error) {
	b.DeleteAllInvocations++
	b.DeleteAllInputs = append(b.DeleteAllInputs, userID)
	if b.DeleteAllStub != nil {
		return b.DeleteAllStub(ctx, userID)
	}
	if len(b.DeleteAllOutputs) > 0 {
		output := b.DeleteAllOutputs[0]
		b.DeleteAllOutputs = b.DeleteAllOutputs[1:]
		return output.Deleted, output.Error
	}
	if b.DeleteAllOutput != nil {
		return b.DeleteAllOutput.Deleted, b.DeleteAllOutput.Error
	}
	panic("DeleteAll has no output")
}

func (b *BlobRepository) DestroyAll(ctx context.Context, userID string) (bool, error) {
	b.DestroyAllInvocations++
	b.DestroyAllInputs = append(b.DestroyAllInputs, userID)
	if b.DestroyAllStub != nil {
		return b.DestroyAllStub(ctx, userID)
	}
	if len(b.DestroyAllOutputs) > 0 {
		output := b.DestroyAllOutputs[0]
		b.DestroyAllOutputs = b.DestroyAllOutputs[1:]
		return output.Destroyed, output.Error
	}
	if b.DestroyAllOutput != nil {
		return b.DestroyAllOutput.Destroyed, b.DestroyAllOutput.Error
	}
	panic("DestroyAll has no output")
}

func (b *BlobRepository) Get(ctx context.Context, id string, condition *request.Condition) (*blob.Blob, error) {
	b.GetInvocations++
	b.GetInputs = append(b.GetInputs, GetInput{ID: id, Condition: condition})
	if b.GetStub != nil {
		return b.GetStub(ctx, id, condition)
	}
	if len(b.GetOutputs) > 0 {
		output := b.GetOutputs[0]
		b.GetOutputs = b.GetOutputs[1:]
		return output.Blob, output.Error
	}
	if b.GetOutput != nil {
		return b.GetOutput.Blob, b.GetOutput.Error
	}
	panic("Get has no output")
}

func (b *BlobRepository) Update(ctx context.Context, id string, condition *request.Condition, update *blobStoreStructured.Update) (*blob.Blob, error) {
	b.UpdateInvocations++
	b.UpdateInputs = append(b.UpdateInputs, UpdateInput{ID: id, Condition: condition, Update: update})
	if b.UpdateStub != nil {
		return b.UpdateStub(ctx, id, condition, update)
	}
	if len(b.UpdateOutputs) > 0 {
		output := b.UpdateOutputs[0]
		b.UpdateOutputs = b.UpdateOutputs[1:]
		return output.Blob, output.Error
	}
	if b.UpdateOutput != nil {
		return b.UpdateOutput.Blob, b.UpdateOutput.Error
	}
	panic("Update has no output")
}

func (b *BlobRepository) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	b.DeleteInvocations++
	b.DeleteInputs = append(b.DeleteInputs, DeleteInput{ID: id, Condition: condition})
	if b.DeleteStub != nil {
		return b.DeleteStub(ctx, id, condition)
	}
	if len(b.DeleteOutputs) > 0 {
		output := b.DeleteOutputs[0]
		b.DeleteOutputs = b.DeleteOutputs[1:]
		return output.Deleted, output.Error
	}
	if b.DeleteOutput != nil {
		return b.DeleteOutput.Deleted, b.DeleteOutput.Error
	}
	panic("Delete has no output")
}

func (b *BlobRepository) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	b.DestroyInvocations++
	b.DestroyInputs = append(b.DestroyInputs, DestroyInput{ID: id, Condition: condition})
	if b.DestroyStub != nil {
		return b.DestroyStub(ctx, id, condition)
	}
	if len(b.DestroyOutputs) > 0 {
		output := b.DestroyOutputs[0]
		b.DestroyOutputs = b.DestroyOutputs[1:]
		return output.Destroyed, output.Error
	}
	if b.DestroyOutput != nil {
		return b.DestroyOutput.Destroyed, b.DestroyOutput.Error
	}
	panic("Destroy has no output")
}

func (b *BlobRepository) AssertOutputsEmpty() {
	if len(b.ListOutputs) > 0 {
		panic("ListOutputs is not empty")
	}
	if len(b.CreateOutputs) > 0 {
		panic("CreateOutputs is not empty")
	}
	if len(b.DeleteAllOutputs) > 0 {
		panic("DeleteAllOutputs is not empty")
	}
	if len(b.DestroyAllOutputs) > 0 {
		panic("DestroyAllOutputs is not empty")
	}
	if len(b.GetOutputs) > 0 {
		panic("GetOutputs is not empty")
	}
	if len(b.UpdateOutputs) > 0 {
		panic("UpdateOutputs is not empty")
	}
	if len(b.DeleteOutputs) > 0 {
		panic("DeleteOutputs is not empty")
	}
	if len(b.DestroyOutputs) > 0 {
		panic("DestroyOutputs is not empty")
	}
}
