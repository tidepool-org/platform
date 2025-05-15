package test

import (
	"context"

	"github.com/tidepool-org/platform/data"
)

type OpenInput struct {
	Context context.Context
	DataSet *data.DataSet
}

type OpenOutput struct {
	DataSet *data.DataSet
	Error   error
}

type AddDataInput struct {
	Context     context.Context
	DataSet     *data.DataSet
	DataSetData data.Data
}

type DeleteDataInput struct {
	Context   context.Context
	DataSet   *data.DataSet
	Selectors *data.Selectors
}

type CloseInput struct {
	Context context.Context
	DataSet *data.DataSet
}

type DeleteInput struct {
	Context context.Context
	DataSet *data.DataSet
}

type Deduplicator struct {
	OpenInvocations       int
	OpenInputs            []OpenInput
	OpenStub              func(ctx context.Context, dataSet *data.DataSet) (*data.DataSet, error)
	OpenOutputs           []OpenOutput
	OpenOutput            *OpenOutput
	AddDataInvocations    int
	AddDataInputs         []AddDataInput
	AddDataStub           func(ctx context.Context, dataSet *data.DataSet, dataSetData data.Data) error
	AddDataOutputs        []error
	AddDataOutput         *error
	DeleteDataInvocations int
	DeleteDataInputs      []DeleteDataInput
	DeleteDataStub        func(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error
	DeleteDataOutputs     []error
	DeleteDataOutput      *error
	CloseInvocations      int
	CloseInputs           []CloseInput
	CloseStub             func(ctx context.Context, dataSet *data.DataSet) error
	CloseOutputs          []error
	CloseOutput           *error
	DeleteInvocations     int
	DeleteInputs          []DeleteInput
	DeleteStub            func(ctx context.Context, dataSet *data.DataSet) error
	DeleteOutputs         []error
	DeleteOutput          *error
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{}
}

func (d *Deduplicator) Open(ctx context.Context, dataSet *data.DataSet) (*data.DataSet, error) {
	d.OpenInvocations++
	d.OpenInputs = append(d.OpenInputs, OpenInput{Context: ctx, DataSet: dataSet})
	if d.OpenStub != nil {
		return d.OpenStub(ctx, dataSet)
	}
	if len(d.OpenOutputs) > 0 {
		output := d.OpenOutputs[0]
		d.OpenOutputs = d.OpenOutputs[1:]
		return output.DataSet, output.Error
	}
	if d.OpenOutput != nil {
		return d.OpenOutput.DataSet, d.OpenOutput.Error
	}
	panic("Open has no output")
}

func (d *Deduplicator) AddData(ctx context.Context, dataSet *data.DataSet, dataSetData data.Data) error {
	d.AddDataInvocations++
	d.AddDataInputs = append(d.AddDataInputs, AddDataInput{Context: ctx, DataSet: dataSet, DataSetData: dataSetData})
	if d.AddDataStub != nil {
		return d.AddDataStub(ctx, dataSet, dataSetData)
	}
	if len(d.AddDataOutputs) > 0 {
		output := d.AddDataOutputs[0]
		d.AddDataOutputs = d.AddDataOutputs[1:]
		return output
	}
	if d.AddDataOutput != nil {
		return *d.AddDataOutput
	}
	panic("AddData has no output")
}

func (d *Deduplicator) DeleteData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	d.DeleteDataInvocations++
	d.DeleteDataInputs = append(d.DeleteDataInputs, DeleteDataInput{Context: ctx, DataSet: dataSet, Selectors: selectors})
	if d.DeleteDataStub != nil {
		return d.DeleteDataStub(ctx, dataSet, selectors)
	}
	if len(d.DeleteDataOutputs) > 0 {
		output := d.DeleteDataOutputs[0]
		d.DeleteDataOutputs = d.DeleteDataOutputs[1:]
		return output
	}
	if d.DeleteDataOutput != nil {
		return *d.DeleteDataOutput
	}
	panic("DeleteData has no output")
}

func (d *Deduplicator) Close(ctx context.Context, dataSet *data.DataSet) error {
	d.CloseInvocations++
	d.CloseInputs = append(d.CloseInputs, CloseInput{Context: ctx, DataSet: dataSet})
	if d.CloseStub != nil {
		return d.CloseStub(ctx, dataSet)
	}
	if len(d.CloseOutputs) > 0 {
		output := d.CloseOutputs[0]
		d.CloseOutputs = d.CloseOutputs[1:]
		return output
	}
	if d.CloseOutput != nil {
		return *d.CloseOutput
	}
	panic("Close has no output")
}

func (d *Deduplicator) Delete(ctx context.Context, dataSet *data.DataSet) error {
	d.DeleteInvocations++
	d.DeleteInputs = append(d.DeleteInputs, DeleteInput{Context: ctx, DataSet: dataSet})
	if d.DeleteStub != nil {
		return d.DeleteStub(ctx, dataSet)
	}
	if len(d.DeleteOutputs) > 0 {
		output := d.DeleteOutputs[0]
		d.DeleteOutputs = d.DeleteOutputs[1:]
		return output
	}
	if d.DeleteOutput != nil {
		return *d.DeleteOutput
	}
	panic("Delete has no output")
}

func (d *Deduplicator) AssertOutputsEmpty() {
	if len(d.OpenOutputs) > 0 {
		panic("OpenOutputs is not empty")
	}
	if len(d.AddDataOutputs) > 0 {
		panic("AddDataOutputs is not empty")
	}
	if len(d.DeleteDataOutputs) > 0 {
		panic("DeleteDataOutputs is not empty")
	}
	if len(d.CloseOutputs) > 0 {
		panic("CloseOutputs is not empty")
	}
	if len(d.DeleteOutputs) > 0 {
		panic("DeleteOutputs is not empty")
	}
}
