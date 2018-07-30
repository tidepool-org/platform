package test

import (
	"context"

	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
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
	Context   context.Context
	ID        string
	Condition *request.Condition
	Update    *dataSource.Update
}

type UpdateOutput struct {
	Source *dataSource.Source
	Error  error
}

type DeleteInput struct {
	Context   context.Context
	ID        string
	Condition *request.Condition
}

type DeleteOutput struct {
	Deleted bool
	Error   error
}

type Session struct {
	*test.Closer
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
	UpdateStub        func(ctx context.Context, id string, condition *request.Condition, create *dataSource.Update) (*dataSource.Source, error)
	UpdateOutputs     []UpdateOutput
	UpdateOutput      *UpdateOutput
	DeleteInvocations int
	DeleteInputs      []DeleteInput
	DeleteStub        func(ctx context.Context, id string, condition *request.Condition) (bool, error)
	DeleteOutputs     []DeleteOutput
	DeleteOutput      *DeleteOutput
}

func NewSession() *Session {
	return &Session{
		Closer: test.NewCloser(),
	}
}

func (s *Session) List(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.Sources, error) {
	s.ListInvocations++
	s.ListInputs = append(s.ListInputs, ListInput{Context: ctx, UserID: userID, Filter: filter, Pagination: pagination})
	if s.ListStub != nil {
		return s.ListStub(ctx, userID, filter, pagination)
	}
	if len(s.ListOutputs) > 0 {
		output := s.ListOutputs[0]
		s.ListOutputs = s.ListOutputs[1:]
		return output.Sources, output.Error
	}
	if s.ListOutput != nil {
		return s.ListOutput.Sources, s.ListOutput.Error
	}
	panic("List has no output")
}

func (s *Session) Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error) {
	s.CreateInvocations++
	s.CreateInputs = append(s.CreateInputs, CreateInput{Context: ctx, UserID: userID, Create: create})
	if s.CreateStub != nil {
		return s.CreateStub(ctx, userID, create)
	}
	if len(s.CreateOutputs) > 0 {
		output := s.CreateOutputs[0]
		s.CreateOutputs = s.CreateOutputs[1:]
		return output.Source, output.Error
	}
	if s.CreateOutput != nil {
		return s.CreateOutput.Source, s.CreateOutput.Error
	}
	panic("Create has no output")
}

func (s *Session) Get(ctx context.Context, id string) (*dataSource.Source, error) {
	s.GetInvocations++
	s.GetInputs = append(s.GetInputs, GetInput{Context: ctx, ID: id})
	if s.GetStub != nil {
		return s.GetStub(ctx, id)
	}
	if len(s.GetOutputs) > 0 {
		output := s.GetOutputs[0]
		s.GetOutputs = s.GetOutputs[1:]
		return output.Source, output.Error
	}
	if s.GetOutput != nil {
		return s.GetOutput.Source, s.GetOutput.Error
	}
	panic("Get has no output")
}

func (s *Session) Update(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
	s.UpdateInvocations++
	s.UpdateInputs = append(s.UpdateInputs, UpdateInput{Context: ctx, ID: id, Condition: condition, Update: update})
	if s.UpdateStub != nil {
		return s.UpdateStub(ctx, id, condition, update)
	}
	if len(s.UpdateOutputs) > 0 {
		output := s.UpdateOutputs[0]
		s.UpdateOutputs = s.UpdateOutputs[1:]
		return output.Source, output.Error
	}
	if s.UpdateOutput != nil {
		return s.UpdateOutput.Source, s.UpdateOutput.Error
	}
	panic("Update has no output")
}

func (s *Session) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	s.DeleteInvocations++
	s.DeleteInputs = append(s.DeleteInputs, DeleteInput{Context: ctx, ID: id, Condition: condition})
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
	if len(s.DeleteOutputs) > 0 {
		panic("DeleteOutputs is not empty")
	}
}
