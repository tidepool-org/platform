package test

import (
	dataDeduplicatorTest "github.com/tidepool-org/platform/data/deduplicator/test"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
)

type NewOutput struct {
	Found bool
	Error error
}

type GetOutput struct {
	Found bool
	Error error
}

type Deduplicator struct {
	*dataDeduplicatorTest.Deduplicator
	NewInvocations int
	NewInputs      []*dataTypesUpload.Upload
	NewStub        func(dataSet *dataTypesUpload.Upload) (bool, error)
	NewOutputs     []NewOutput
	NewOutput      *NewOutput
	GetInvocations int
	GetInputs      []*dataTypesUpload.Upload
	GetStub        func(dataSet *dataTypesUpload.Upload) (bool, error)
	GetOutputs     []GetOutput
	GetOutput      *GetOutput
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		Deduplicator: dataDeduplicatorTest.NewDeduplicator(),
	}
}

func (d *Deduplicator) New(dataSet *dataTypesUpload.Upload) (bool, error) {
	d.NewInvocations++
	d.NewInputs = append(d.NewInputs, dataSet)
	if d.NewStub != nil {
		return d.NewStub(dataSet)
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

func (d *Deduplicator) Get(dataSet *dataTypesUpload.Upload) (bool, error) {
	d.GetInvocations++
	d.GetInputs = append(d.GetInputs, dataSet)
	if d.GetStub != nil {
		return d.GetStub(dataSet)
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
