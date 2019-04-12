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

type Session struct {
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

func NewSession() *Session {
	return &Session{
		Closer: test.NewCloser(),
	}
}

func (s *Session) List(ctx context.Context, userID string, filter *image.Filter, pagination *page.Pagination) (image.ImageArray, error) {
	s.ListInvocations++
	s.ListInputs = append(s.ListInputs, ListInput{UserID: userID, Filter: filter, Pagination: pagination})
	if s.ListStub != nil {
		return s.ListStub(ctx, userID, filter, pagination)
	}
	if len(s.ListOutputs) > 0 {
		output := s.ListOutputs[0]
		s.ListOutputs = s.ListOutputs[1:]
		return output.ImageArray, output.Error
	}
	if s.ListOutput != nil {
		return s.ListOutput.ImageArray, s.ListOutput.Error
	}
	panic("List has no output")
}

func (s *Session) Create(ctx context.Context, userID string, metadata *image.Metadata) (*image.Image, error) {
	s.CreateInvocations++
	s.CreateInputs = append(s.CreateInputs, CreateInput{UserID: userID, Metadata: metadata})
	if s.CreateStub != nil {
		return s.CreateStub(ctx, userID, metadata)
	}
	if len(s.CreateOutputs) > 0 {
		output := s.CreateOutputs[0]
		s.CreateOutputs = s.CreateOutputs[1:]
		return output.Image, output.Error
	}
	if s.CreateOutput != nil {
		return s.CreateOutput.Image, s.CreateOutput.Error
	}
	panic("Create has no output")
}

func (s *Session) DeleteAll(ctx context.Context, userID string) (bool, error) {
	s.DeleteAllInvocations++
	s.DeleteAllInputs = append(s.DeleteAllInputs, userID)
	if s.DeleteAllStub != nil {
		return s.DeleteAllStub(ctx, userID)
	}
	if len(s.DeleteAllOutputs) > 0 {
		output := s.DeleteAllOutputs[0]
		s.DeleteAllOutputs = s.DeleteAllOutputs[1:]
		return output.Deleted, output.Error
	}
	if s.DeleteAllOutput != nil {
		return s.DeleteAllOutput.Deleted, s.DeleteAllOutput.Error
	}
	panic("DeleteAll has no output")
}

func (s *Session) DestroyAll(ctx context.Context, userID string) (bool, error) {
	s.DestroyAllInvocations++
	s.DestroyAllInputs = append(s.DestroyAllInputs, userID)
	if s.DestroyAllStub != nil {
		return s.DestroyAllStub(ctx, userID)
	}
	if len(s.DestroyAllOutputs) > 0 {
		output := s.DestroyAllOutputs[0]
		s.DestroyAllOutputs = s.DestroyAllOutputs[1:]
		return output.Destroyed, output.Error
	}
	if s.DestroyAllOutput != nil {
		return s.DestroyAllOutput.Destroyed, s.DestroyAllOutput.Error
	}
	panic("DestroyAll has no output")
}

func (s *Session) Get(ctx context.Context, id string, condition *request.Condition) (*image.Image, error) {
	s.GetInvocations++
	s.GetInputs = append(s.GetInputs, GetInput{ID: id, Condition: condition})
	if s.GetStub != nil {
		return s.GetStub(ctx, id, condition)
	}
	if len(s.GetOutputs) > 0 {
		output := s.GetOutputs[0]
		s.GetOutputs = s.GetOutputs[1:]
		return output.Image, output.Error
	}
	if s.GetOutput != nil {
		return s.GetOutput.Image, s.GetOutput.Error
	}
	panic("Get has no output")
}

func (s *Session) Update(ctx context.Context, id string, condition *request.Condition, update *imageStoreStructured.Update) (*image.Image, error) {
	s.UpdateInvocations++
	s.UpdateInputs = append(s.UpdateInputs, UpdateInput{ID: id, Condition: condition, Update: update})
	if s.UpdateStub != nil {
		return s.UpdateStub(ctx, id, condition, update)
	}
	if len(s.UpdateOutputs) > 0 {
		output := s.UpdateOutputs[0]
		s.UpdateOutputs = s.UpdateOutputs[1:]
		return output.Image, output.Error
	}
	if s.UpdateOutput != nil {
		return s.UpdateOutput.Image, s.UpdateOutput.Error
	}
	panic("Update has no output")
}

func (s *Session) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	s.DeleteInvocations++
	s.DeleteInputs = append(s.DeleteInputs, DeleteInput{ID: id, Condition: condition})
	if s.DeleteStub != nil {
		return s.DeleteStub(ctx, id, condition)
	}
	if len(s.DeleteOutputs) > 0 {
		output := s.DeleteOutputs[0]
		s.DeleteOutputs = s.DeleteOutputs[1:]
		return output.Deleted, output.Error
	}
	if s.DeleteOutput != nil {
		return s.DeleteOutput.Deleted, s.DeleteOutput.Error
	}
	panic("Delete has no output")
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
	if len(s.DeleteAllOutputs) > 0 {
		panic("DeleteAllOutputs is not empty")
	}
	if len(s.DestroyAllOutputs) > 0 {
		panic("DestroyAllOutputs is not empty")
	}
	if len(s.GetOutputs) > 0 {
		panic("GetOutputs is not empty")
	}
	if len(s.UpdateOutputs) > 0 {
		panic("UpdateOutputs is not empty")
	}
	if len(s.DeleteOutputs) > 0 {
		panic("DeleteOutputs is not empty")
	}
	if len(s.DestroyOutputs) > 0 {
		panic("DestroyOutputs is not empty")
	}
}
