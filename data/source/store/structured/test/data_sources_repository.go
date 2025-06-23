package test

import (
	"context"

	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
)

type ListInput struct {
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
	ListStub              func(ctx context.Context, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error)
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

func (d *DataRepository) List(ctx context.Context, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error) {
	d.ListInvocations++
	d.ListInputs = append(d.ListInputs, ListInput{Filter: filter, Pagination: pagination})
	if d.ListStub != nil {
		return d.ListStub(ctx, filter, pagination)
	}
	if len(d.ListOutputs) > 0 {
		output := d.ListOutputs[0]
		d.ListOutputs = d.ListOutputs[1:]
		return output.SourceArray, output.Error
	}
	if d.ListOutput != nil {
		return d.ListOutput.SourceArray, d.ListOutput.Error
	}
	panic("List has no output")
}

func (d *DataRepository) Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error) {
	d.CreateInvocations++
	d.CreateInputs = append(d.CreateInputs, CreateInput{UserID: userID, Create: create})
	if d.CreateStub != nil {
		return d.CreateStub(ctx, userID, create)
	}
	if len(d.CreateOutputs) > 0 {
		output := d.CreateOutputs[0]
		d.CreateOutputs = d.CreateOutputs[1:]
		return output.Source, output.Error
	}
	if d.CreateOutput != nil {
		return d.CreateOutput.Source, d.CreateOutput.Error
	}
	panic("Create has no output")
}

func (d *DataRepository) DestroyAll(ctx context.Context, userID string) (bool, error) {
	d.DestroyAllInvocations++
	d.DestroyAllInputs = append(d.DestroyAllInputs, userID)
	if d.DestroyAllStub != nil {
		return d.DestroyAllStub(ctx, userID)
	}
	if len(d.DestroyAllOutputs) > 0 {
		output := d.DestroyAllOutputs[0]
		d.DestroyAllOutputs = d.DestroyAllOutputs[1:]
		return output.Destroyed, output.Error
	}
	if d.DestroyAllOutput != nil {
		return d.DestroyAllOutput.Destroyed, d.DestroyAllOutput.Error
	}
	panic("DestroyAll has no output")
}

func (d *DataRepository) Get(ctx context.Context, id string) (*dataSource.Source, error) {
	d.GetInvocations++
	d.GetInputs = append(d.GetInputs, id)
	if d.GetStub != nil {
		return d.GetStub(ctx, id)
	}
	if len(d.GetOutputs) > 0 {
		output := d.GetOutputs[0]
		d.GetOutputs = d.GetOutputs[1:]
		return output.Source, output.Error
	}
	if d.GetOutput != nil {
		return d.GetOutput.Source, d.GetOutput.Error
	}
	panic("Get has no output")
}

func (d *DataRepository) Update(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
	d.UpdateInvocations++
	d.UpdateInputs = append(d.UpdateInputs, UpdateInput{ID: id, Condition: condition, Update: update})
	if d.UpdateStub != nil {
		return d.UpdateStub(ctx, id, condition, update)
	}
	if len(d.UpdateOutputs) > 0 {
		output := d.UpdateOutputs[0]
		d.UpdateOutputs = d.UpdateOutputs[1:]
		return output.Source, output.Error
	}
	if d.UpdateOutput != nil {
		return d.UpdateOutput.Source, d.UpdateOutput.Error
	}
	panic("Update has no output")
}

func (d *DataRepository) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	d.DestroyInvocations++
	d.DestroyInputs = append(d.DestroyInputs, DestroyInput{ID: id, Condition: condition})
	if d.DestroyStub != nil {
		return d.DestroyStub(ctx, id, condition)
	}
	if len(d.DestroyOutputs) > 0 {
		output := d.DestroyOutputs[0]
		d.DestroyOutputs = d.DestroyOutputs[1:]
		return output.Destroyed, output.Error
	}
	if d.DestroyOutput != nil {
		return d.DestroyOutput.Destroyed, d.DestroyOutput.Error
	}
	panic("Destroy has no output")
}

func (d *DataRepository) AssertOutputsEmpty() {
	d.Closer.AssertOutputsEmpty()
	if len(d.ListOutputs) > 0 {
		panic("ListOutputs is not empty")
	}
	if len(d.CreateOutputs) > 0 {
		panic("CreateOutputs is not empty")
	}
	if len(d.DestroyAllOutputs) > 0 {
		panic("DestroyAllOutputs is not empty")
	}
	if len(d.GetOutputs) > 0 {
		panic("GetOutputs is not empty")
	}
	if len(d.UpdateOutputs) > 0 {
		panic("UpdateOutputs is not empty")
	}
	if len(d.DestroyOutputs) > 0 {
		panic("DestroyOutputs is not empty")
	}
}
