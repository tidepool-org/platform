package test

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
)

type OpenInput struct {
	Context    context.Context
	Repository dataStore.DataRepository
	DataSet    *dataTypesUpload.Upload
}

type OpenOutput struct {
	DataSet *dataTypesUpload.Upload
	Error   error
}

type AddDataInput struct {
	Context     context.Context
	Repository  dataStore.DataRepository
	DataSet     *dataTypesUpload.Upload
	DataSetData data.Data
}

type DeleteDataInput struct {
	Context    context.Context
	Repository dataStore.DataRepository
	DataSet    *dataTypesUpload.Upload
	Selectors  *data.Selectors
}

type CloseInput struct {
	Context    context.Context
	Repository dataStore.DataRepository
	DataSet    *dataTypesUpload.Upload
}

type DeleteInput struct {
	Context    context.Context
	Repository dataStore.DataRepository
	DataSet    *dataTypesUpload.Upload
	doPurge    bool
}

type Deduplicator struct {
	OpenInvocations       int
	OpenInputs            []OpenInput
	OpenStub              func(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload) (*dataTypesUpload.Upload, error)
	OpenOutputs           []OpenOutput
	OpenOutput            *OpenOutput
	AddDataInvocations    int
	AddDataInputs         []AddDataInput
	AddDataStub           func(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload, dataSetData data.Data) error
	AddDataOutputs        []error
	AddDataOutput         *error
	DeleteDataInvocations int
	DeleteDataInputs      []DeleteDataInput
	DeleteDataStub        func(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload, selectors *data.Selectors) error
	DeleteDataOutputs     []error
	DeleteDataOutput      *error
	CloseInvocations      int
	CloseInputs           []CloseInput
	CloseStub             func(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload) error
	CloseOutputs          []error
	CloseOutput           *error
	DeleteInvocations     int
	DeleteInputs          []DeleteInput
	DeleteStub            func(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload, doPurge bool) error
	DeleteOutputs         []error
	DeleteOutput          *error
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{}
}

func (d *Deduplicator) Open(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload) (*dataTypesUpload.Upload, error) {
	d.OpenInvocations++
	d.OpenInputs = append(d.OpenInputs, OpenInput{Context: ctx, Repository: repository, DataSet: dataSet})
	if d.OpenStub != nil {
		return d.OpenStub(ctx, repository, dataSet)
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

func (d *Deduplicator) AddData(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload, dataSetData data.Data) error {
	d.AddDataInvocations++
	d.AddDataInputs = append(d.AddDataInputs, AddDataInput{Context: ctx, Repository: repository, DataSet: dataSet, DataSetData: dataSetData})
	if d.AddDataStub != nil {
		return d.AddDataStub(ctx, repository, dataSet, dataSetData)
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

func (d *Deduplicator) DeleteData(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload, selectors *data.Selectors) error {
	d.DeleteDataInvocations++
	d.DeleteDataInputs = append(d.DeleteDataInputs, DeleteDataInput{Context: ctx, Repository: repository, DataSet: dataSet, Selectors: selectors})
	if d.DeleteDataStub != nil {
		return d.DeleteDataStub(ctx, repository, dataSet, selectors)
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

func (d *Deduplicator) Close(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload) error {
	d.CloseInvocations++
	d.CloseInputs = append(d.CloseInputs, CloseInput{Context: ctx, Repository: repository, DataSet: dataSet})
	if d.CloseStub != nil {
		return d.CloseStub(ctx, repository, dataSet)
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

func (d *Deduplicator) Delete(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload, doPurge bool) error {
	d.DeleteInvocations++
	d.DeleteInputs = append(d.DeleteInputs, DeleteInput{Context: ctx, Repository: repository, DataSet: dataSet})
	if d.DeleteStub != nil {
		return d.DeleteStub(ctx, repository, dataSet, doPurge)
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
