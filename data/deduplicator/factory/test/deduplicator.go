package test

import (
	"context"

	dataDeduplicatorTest "github.com/tidepool-org/platform/data/deduplicator/test"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
)

type NewInput struct {
	Context context.Context
	DataSet *dataTypesUpload.Upload
}

type NewOutput struct {
	Found bool
	Error error
}

type GetInput struct {
	Context context.Context
	DataSet *dataTypesUpload.Upload
}

type GetOutput struct {
	Found bool
	Error error
}

type Deduplicator struct {
	*dataDeduplicatorTest.Deduplicator
	NewInvocations int
	NewInputs      []NewInput
	NewStub        func(ctx context.Context, dataSet *dataTypesUpload.Upload) (bool, error)
	NewOutputs     []NewOutput
	NewOutput      *NewOutput
	GetInvocations int
	GetInputs      []GetInput
	GetStub        func(ctx context.Context, dataSet *dataTypesUpload.Upload) (bool, error)
	GetOutputs     []GetOutput
	GetOutput      *GetOutput
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		Deduplicator: dataDeduplicatorTest.NewDeduplicator(),
	}
}

func (d *Deduplicator) New(ctx context.Context, dataSet *dataTypesUpload.Upload) (bool, error) {
	d.NewInvocations++
	d.NewInputs = append(d.NewInputs, NewInput{Context: ctx, DataSet: dataSet})
	if d.NewStub != nil {
		return d.NewStub(ctx, dataSet)
	}
	if len(d.NewOutputs) > 0 {
		output := d.NewOutputs[0]
		d.NewOutputs = d.NewOutputs[1:]
		return output.Found, output.Error
	}
	if d.NewOutput != nil {
		return d.NewOutput.Found, d.NewOutput.Error
	}
	panic("New has no output")
}

func (d *Deduplicator) Get(ctx context.Context, dataSet *dataTypesUpload.Upload) (bool, error) {
	d.GetInvocations++
	d.GetInputs = append(d.GetInputs, GetInput{Context: ctx, DataSet: dataSet})
	if d.GetStub != nil {
		return d.GetStub(ctx, dataSet)
	}
	if len(d.GetOutputs) > 0 {
		output := d.GetOutputs[0]
		d.GetOutputs = d.GetOutputs[1:]
		return output.Found, output.Error
	}
	if d.GetOutput != nil {
		return d.GetOutput.Found, d.GetOutput.Error
	}
	panic("Get has no output")
}

func (d *Deduplicator) AssertOutputsEmpty() {
	d.Deduplicator.AssertOutputsEmpty()
	if len(d.NewOutputs) > 0 {
		panic("NewOutputs is not empty")
	}
	if len(d.GetOutputs) > 0 {
		panic("GetOutputs is not empty")
	}
}
