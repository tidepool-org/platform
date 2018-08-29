package test

import (
	dataDeduplicator "github.com/tidepool-org/platform/data/deduplicator"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
)

type NewOutput struct {
	Deduplicator dataDeduplicator.Deduplicator
	Error        error
}

type GetOutput struct {
	Deduplicator dataDeduplicator.Deduplicator
	Error        error
}

type Factory struct {
	NewInvocations int
	NewInputs      []*dataTypesUpload.Upload
	NewStub        func(dataSet *dataTypesUpload.Upload) (dataDeduplicator.Deduplicator, error)
	NewOutputs     []NewOutput
	NewOutput      *NewOutput
	GetInvocations int
	GetInputs      []*dataTypesUpload.Upload
	GetStub        func(dataSet *dataTypesUpload.Upload) (dataDeduplicator.Deduplicator, error)
	GetOutputs     []GetOutput
	GetOutput      *GetOutput
}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) New(dataSet *dataTypesUpload.Upload) (dataDeduplicator.Deduplicator, error) {
	f.NewInvocations++
	f.NewInputs = append(f.NewInputs, dataSet)
	if f.NewStub != nil {
		return f.NewStub(dataSet)
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

func (f *Factory) Get(dataSet *dataTypesUpload.Upload) (dataDeduplicator.Deduplicator, error) {
	f.GetInvocations++
	f.GetInputs = append(f.GetInputs, dataSet)
	if f.GetStub != nil {
		return f.GetStub(dataSet)
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
