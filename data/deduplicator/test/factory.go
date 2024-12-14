package test

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataDeduplicator "github.com/tidepool-org/platform/data/deduplicator"
)

type NewInput struct {
	Context context.Context
	DataSet *data.DataSet
}

type NewOutput struct {
	Deduplicator dataDeduplicator.Deduplicator
	Error        error
}

type GetInput struct {
	Context context.Context
	DataSet *data.DataSet
}

type GetOutput struct {
	Deduplicator dataDeduplicator.Deduplicator
	Error        error
}

type Factory struct {
	NewInvocations int
	NewInputs      []NewInput
	NewStub        func(ctx context.Context, dataSet *data.DataSet) (dataDeduplicator.Deduplicator, error)
	NewOutputs     []NewOutput
	NewOutput      *NewOutput
	GetInvocations int
	GetInputs      []GetInput
	GetStub        func(ctx context.Context, dataSet *data.DataSet) (dataDeduplicator.Deduplicator, error)
	GetOutputs     []GetOutput
	GetOutput      *GetOutput
}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) New(ctx context.Context, dataSet *data.DataSet) (dataDeduplicator.Deduplicator, error) {
	f.NewInvocations++
	f.NewInputs = append(f.NewInputs, NewInput{Context: ctx, DataSet: dataSet})
	if f.NewStub != nil {
		return f.NewStub(ctx, dataSet)
	}
	if len(f.NewOutputs) > 0 {
		output := f.NewOutputs[0]
		f.NewOutputs = f.NewOutputs[1:]
		return output.Deduplicator, output.Error
	}
	if f.NewOutput != nil {
		return f.NewOutput.Deduplicator, f.NewOutput.Error
	}
	panic("New has no output")
}

func (f *Factory) Get(ctx context.Context, dataSet *data.DataSet) (dataDeduplicator.Deduplicator, error) {
	f.GetInvocations++
	f.GetInputs = append(f.GetInputs, GetInput{Context: ctx, DataSet: dataSet})
	if f.GetStub != nil {
		return f.GetStub(ctx, dataSet)
	}
	if len(f.GetOutputs) > 0 {
		output := f.GetOutputs[0]
		f.GetOutputs = f.GetOutputs[1:]
		return output.Deduplicator, output.Error
	}
	if f.GetOutput != nil {
		return f.GetOutput.Deduplicator, f.GetOutput.Error
	}
	panic("Get has no output")
}

func (f *Factory) AssertOutputsEmpty() {
	if len(f.NewOutputs) > 0 {
		panic("NewOutputs is not empty")
	}
	if len(f.GetOutputs) > 0 {
		panic("GetOutputs is not empty")
	}
}
