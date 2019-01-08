package test

import (
	"context"

	"github.com/tidepool-org/platform/blob"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
)

type ListInput struct {
	UserID     string
	Filter     *blob.Filter
	Pagination *page.Pagination
}

type ListOutput struct {
	Blobs blob.Blobs
	Error error
}

type CreateInput struct {
	UserID string
	Create *blobStoreStructured.Create
}

type CreateOutput struct {
	Blob  *blob.Blob
	Error error
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

type DestroyInput struct {
	ID        string
	Condition *request.Condition
}

type DestroyOutput struct {
	Destroyed bool
	Error     error
}

type Session struct {
	*test.Closer
	ListInvocations    int
	ListInputs         []ListInput
	ListStub           func(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.Blobs, error)
	ListOutputs        []ListOutput
	ListOutput         *ListOutput
	CreateInvocations  int
	CreateInputs       []CreateInput
	CreateStub         func(ctx context.Context, userID string, create *blobStoreStructured.Create) (*blob.Blob, error)
	CreateOutputs      []CreateOutput
	CreateOutput       *CreateOutput
	GetInvocations     int
	GetInputs          []string
	GetStub            func(ctx context.Context, id string) (*blob.Blob, error)
	GetOutputs         []GetOutput
	GetOutput          *GetOutput
	UpdateInvocations  int
	UpdateInputs       []UpdateInput
	UpdateStub         func(ctx context.Context, id string, condition *request.Condition, update *blobStoreStructured.Update) (*blob.Blob, error)
	UpdateOutputs      []UpdateOutput
	UpdateOutput       *UpdateOutput
	DestroyInvocations int
	DestroyInputs      []DestroyInput
	DestroyStub        func(ctx context.Context, id string, condition *request.Condition) (bool, error)
	DestroyOutputs     []DestroyOutput
	DestroyOutput      *DestroyOutput
}

func NewSession() *Session {
	return &Session{
		Closer: test.NewCloser(),
	}
}

func (s *Session) List(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.Blobs, error) {
	s.ListInvocations++
	s.ListInputs = append(s.ListInputs, ListInput{UserID: userID, Filter: filter, Pagination: pagination})
	if s.ListStub != nil {
		return s.ListStub(ctx, userID, filter, pagination)
	}
	if len(s.ListOutputs) > 0 {
		output := s.ListOutputs[0]
		s.ListOutputs = s.ListOutputs[1:]
		return output.Blobs, output.Error
	}
	if s.ListOutput != nil {
		return s.ListOutput.Blobs, s.ListOutput.Error
	}
	panic("List has no output")
}

func (s *Session) Create(ctx context.Context, userID string, create *blobStoreStructured.Create) (*blob.Blob, error) {
	s.CreateInvocations++
	s.CreateInputs = append(s.CreateInputs, CreateInput{UserID: userID, Create: create})
	if s.CreateStub != nil {
		return s.CreateStub(ctx, userID, create)
	}
	if len(s.CreateOutputs) > 0 {
		output := s.CreateOutputs[0]
		s.CreateOutputs = s.CreateOutputs[1:]
		return output.Blob, output.Error
	}
	if s.CreateOutput != nil {
		return s.CreateOutput.Blob, s.CreateOutput.Error
	}
	panic("Create has no output")
}

func (s *Session) Get(ctx context.Context, id string) (*blob.Blob, error) {
	s.GetInvocations++
	s.GetInputs = append(s.GetInputs, id)
	if s.GetStub != nil {
		return s.GetStub(ctx, id)
	}
	if len(s.GetOutputs) > 0 {
		output := s.GetOutputs[0]
		s.GetOutputs = s.GetOutputs[1:]
		return output.Blob, output.Error
	}
	if s.GetOutput != nil {
		return s.GetOutput.Blob, s.GetOutput.Error
	}
	panic("Get has no output")
}

func (s *Session) Update(ctx context.Context, id string, condition *request.Condition, update *blobStoreStructured.Update) (*blob.Blob, error) {
	s.UpdateInvocations++
	s.UpdateInputs = append(s.UpdateInputs, UpdateInput{ID: id, Condition: condition, Update: update})
	if s.UpdateStub != nil {
		return s.UpdateStub(ctx, id, condition, update)
	}
	if len(s.UpdateOutputs) > 0 {
		output := s.UpdateOutputs[0]
		s.UpdateOutputs = s.UpdateOutputs[1:]
		return output.Blob, output.Error
	}
	if s.UpdateOutput != nil {
		return s.UpdateOutput.Blob, s.UpdateOutput.Error
	}
	panic("Update has no output")
}

func (s *Session) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	s.DestroyInvocations++
	s.DestroyInputs = append(s.DestroyInputs, DestroyInput{ID: id, Condition: condition})
	if s.DestroyStub != nil {
		return s.DestroyStub(ctx, id, condition)
	}
	if len(s.DestroyOutputs) > 0 {
		output := s.DestroyOutputs[0]
		s.DestroyOutputs = s.DestroyOutputs[1:]
		return output.Destroyed, output.Error
	}
	if s.DestroyOutput != nil {
		return s.DestroyOutput.Destroyed, s.DestroyOutput.Error
	}
	panic("Destroy has no output")
}

func (s *Session) AssertOutputsEmpty() {
	s.Closer.AssertOutputsEmpty()
	if len(s.ListOutputs) > 0 {
		panic("ListOutputs is not empty")
	}
	if len(s.CreateOutputs) > 0 {
		panic("CreateOutputs is not empty")
	}
	if len(s.GetOutputs) > 0 {
		panic("GetOutputs is not empty")
	}
	if len(s.UpdateOutputs) > 0 {
		panic("UpdateOutputs is not empty")
	}
	if len(s.DestroyOutputs) > 0 {
		panic("DestroyOutputs is not empty")
	}
}
