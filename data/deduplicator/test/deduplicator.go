package test

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
)

type OpenInput struct {
	Context context.Context
	Session dataStoreDEPRECATED.DataSession
	DataSet *dataTypesUpload.Upload
}

type OpenOutput struct {
	DataSet *dataTypesUpload.Upload
	Error   error
}

type AddDataInput struct {
	Context     context.Context
	Session     dataStoreDEPRECATED.DataSession
	DataSet     *dataTypesUpload.Upload
	DataSetData data.Data
}

type DeleteDataInput struct {
	Context context.Context
	Session dataStoreDEPRECATED.DataSession
	DataSet *dataTypesUpload.Upload
	Deletes *data.Deletes
}

type CloseInput struct {
	Context context.Context
	Session dataStoreDEPRECATED.DataSession
	DataSet *dataTypesUpload.Upload
}

type DeleteInput struct {
	Context context.Context
	Session dataStoreDEPRECATED.DataSession
	DataSet *dataTypesUpload.Upload
}

type Deduplicator struct {
	OpenInvocations       int
	OpenInputs            []OpenInput
	OpenStub              func(ctx context.Context, session dataStoreDEPRECATED.DataSession, dataSet *dataTypesUpload.Upload) (*dataTypesUpload.Upload, error)
	OpenOutputs           []OpenOutput
	OpenOutput            *OpenOutput
	AddDataInvocations    int
	AddDataInputs         []AddDataInput
	AddDataStub           func(ctx context.Context, session dataStoreDEPRECATED.DataSession, dataSet *dataTypesUpload.Upload, dataSetData data.Data) error
	AddDataOutputs        []error
	AddDataOutput         *error
	DeleteDataInvocations int
	DeleteDataInputs      []DeleteDataInput
	DeleteDataStub        func(ctx context.Context, session dataStoreDEPRECATED.DataSession, dataSet *dataTypesUpload.Upload, deletes *data.Deletes) error
	DeleteDataOutputs     []error
	DeleteDataOutput      *error
	CloseInvocations      int
	CloseInputs           []CloseInput
	CloseStub             func(ctx context.Context, session dataStoreDEPRECATED.DataSession, dataSet *dataTypesUpload.Upload) error
	CloseOutputs          []error
	CloseOutput           *error
	DeleteInvocations     int
	DeleteInputs          []DeleteInput
	DeleteStub            func(ctx context.Context, session dataStoreDEPRECATED.DataSession, dataSet *dataTypesUpload.Upload) error
	DeleteOutputs         []error
	DeleteOutput          *error
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{}
}

func (d *Deduplicator) Open(ctx context.Context, session dataStoreDEPRECATED.DataSession, dataSet *dataTypesUpload.Upload) (*dataTypesUpload.Upload, error) {
	d.OpenInvocations++
	d.OpenInputs = append(d.OpenInputs, OpenInput{Context: ctx, Session: session, DataSet: dataSet})
	if d.OpenStub != nil {
		return d.OpenStub(ctx, session, dataSet)
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

func (d *Deduplicator) AddData(ctx context.Context, session dataStoreDEPRECATED.DataSession, dataSet *dataTypesUpload.Upload, dataSetData data.Data) error {
	d.AddDataInvocations++
	d.AddDataInputs = append(d.AddDataInputs, AddDataInput{Context: ctx, Session: session, DataSet: dataSet, DataSetData: dataSetData})
	if d.AddDataStub != nil {
		return d.AddDataStub(ctx, session, dataSet, dataSetData)
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

func (d *Deduplicator) DeleteData(ctx context.Context, session dataStoreDEPRECATED.DataSession, dataSet *dataTypesUpload.Upload, deletes *data.Deletes) error {
	d.DeleteDataInvocations++
	d.DeleteDataInputs = append(d.DeleteDataInputs, DeleteDataInput{Context: ctx, Session: session, DataSet: dataSet, Deletes: deletes})
	if d.DeleteDataStub != nil {
		return d.DeleteDataStub(ctx, session, dataSet, deletes)
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

func (d *Deduplicator) Close(ctx context.Context, session dataStoreDEPRECATED.DataSession, dataSet *dataTypesUpload.Upload) error {
	d.CloseInvocations++
	d.CloseInputs = append(d.CloseInputs, CloseInput{Context: ctx, Session: session, DataSet: dataSet})
	if d.CloseStub != nil {
		return d.CloseStub(ctx, session, dataSet)
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

func (d *Deduplicator) Delete(ctx context.Context, session dataStoreDEPRECATED.DataSession, dataSet *dataTypesUpload.Upload) error {
	d.DeleteInvocations++
	d.DeleteInputs = append(d.DeleteInputs, DeleteInput{Context: ctx, Session: session, DataSet: dataSet})
	if d.DeleteStub != nil {
		return d.DeleteStub(ctx, session, dataSet)
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
