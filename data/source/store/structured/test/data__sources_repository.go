package test

import (
	"context"

	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
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

type DestroyAllOutput struct {
	Destroyed bool
	Error     error
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

type DestroyInput struct {
	ID        string
	Condition *request.Condition
}

type DestroyOutput struct {
	Destroyed bool
	Error     error
}

type DataRepository struct {
	*test.Closer
	ListInvocations       int
	ListInputs            []ListInput
	ListStub              func(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error)
	ListOutputs           []ListOutput
	ListOutput            *ListOutput
	CreateInvocations     int
	CreateInputs          []CreateInput
	CreateStub            func(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error)
	CreateOutputs         []CreateOutput
	CreateOutput          *CreateOutput
	DestroyAllInvocations int
	DestroyAllInputs      []string
	DestroyAllStub        func(ctx context.Context, userID string) (bool, error)
	DestroyAllOutputs     []DestroyAllOutput
	DestroyAllOutput      *DestroyAllOutput
	GetInvocations        int
	GetInputs             []string
	GetStub               func(ctx context.Context, id string) (*dataSource.Source, error)
	GetOutputs            []GetOutput
	GetOutput             *GetOutput
	UpdateInvocations     int
	UpdateInputs          []UpdateInput
	UpdateStub            func(ctx context.Context, id string, condition *request.Condition, create *dataSource.Update) (*dataSource.Source, error)
	UpdateOutputs         []UpdateOutput
	UpdateOutput          *UpdateOutput
	DestroyInvocations    int
	DestroyInputs         []DestroyInput
	DestroyStub           func(ctx context.Context, id string, condition *request.Condition) (bool, error)
	DestroyOutputs        []DestroyOutput
	DestroyOutput         *DestroyOutput
}

func NewDataSourcesRepository() *DataRepository {
	return &DataRepository{
		Closer: test.NewCloser(),
	}
}

func (s *DataRepository) List(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error) {
	s.ListInvocations++
	s.ListInputs = append(s.ListInputs, ListInput{UserID: userID, Filter: filter, Pagination: pagination})
	if s.ListStub != nil {
		return s.ListStub(ctx, userID, filter, pagination)
	}
	if len(s.ListOutputs) > 0 {
		output := s.ListOutputs[0]
		s.ListOutputs = s.ListOutputs[1:]
		return output.SourceArray, output.Error
	}
	if s.ListOutput != nil {
		return s.ListOutput.SourceArray, s.ListOutput.Error
	}
	panic("List has no output")
}

func (s *DataRepository) Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error) {
	s.CreateInvocations++
	s.CreateInputs = append(s.CreateInputs, CreateInput{UserID: userID, Create: create})
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

func (s *DataRepository) DestroyAll(ctx context.Context, userID string) (bool, error) {
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

func (s *DataRepository) Get(ctx context.Context, id string) (*dataSource.Source, error) {
	s.GetInvocations++
	s.GetInputs = append(s.GetInputs, id)
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

func (s *DataRepository) Update(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
	s.UpdateInvocations++
	s.UpdateInputs = append(s.UpdateInputs, UpdateInput{ID: id, Condition: condition, Update: update})
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

func (s *DataRepository) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
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

func (s *DataRepository) AssertOutputsEmpty() {
	s.Closer.AssertOutputsEmpty()
	if len(s.ListOutputs) > 0 {
		panic("ListOutputs is not empty")
	}
	if len(s.CreateOutputs) > 0 {
		panic("CreateOutputs is not empty")
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
	if len(s.DestroyOutputs) > 0 {
		panic("DestroyOutputs is not empty")
	}
}
