package test

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/test"
)

type AddDataSetDataInput struct {
	Context     context.Context
	DataSetData []data.Datum
}

type Deduplicator struct {
	*test.Mock
	NameInvocations               int
	NameOutputs                   []string
	VersionInvocations            int
	VersionOutputs                []string
	RegisterDataSetInvocations    int
	RegisterDataSetInputs         []context.Context
	RegisterDataSetOutputs        []error
	AddDataSetDataInvocations     int
	AddDataSetDataInputs          []AddDataSetDataInput
	AddDataSetDataOutputs         []error
	DeduplicateDataSetInvocations int
	DeduplicateDataSetInputs      []context.Context
	DeduplicateDataSetOutputs     []error
	DeleteDataSetInvocations      int
	DeleteDataSetInputs           []context.Context
	DeleteDataSetOutputs          []error
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		Mock: test.NewMock(),
	}
}

func (d *Deduplicator) Name() string {
	d.NameInvocations++

	gomega.Expect(d.NameOutputs).ToNot(gomega.BeEmpty())

	output := d.NameOutputs[0]
	d.NameOutputs = d.NameOutputs[1:]
	return output
}

func (d *Deduplicator) Version() string {
	d.VersionInvocations++

	gomega.Expect(d.VersionOutputs).ToNot(gomega.BeEmpty())

	output := d.VersionOutputs[0]
	d.VersionOutputs = d.VersionOutputs[1:]
	return output
}

func (d *Deduplicator) RegisterDataSet(ctx context.Context) error {
	d.RegisterDataSetInvocations++

	d.RegisterDataSetInputs = append(d.RegisterDataSetInputs, ctx)

	gomega.Expect(d.RegisterDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.RegisterDataSetOutputs[0]
	d.RegisterDataSetOutputs = d.RegisterDataSetOutputs[1:]
	return output
}

func (d *Deduplicator) AddDataSetData(ctx context.Context, dataSetData []data.Datum) error {
	d.AddDataSetDataInvocations++

	d.AddDataSetDataInputs = append(d.AddDataSetDataInputs, AddDataSetDataInput{Context: ctx, DataSetData: dataSetData})

	gomega.Expect(d.AddDataSetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.AddDataSetDataOutputs[0]
	d.AddDataSetDataOutputs = d.AddDataSetDataOutputs[1:]
	return output
}

func (d *Deduplicator) DeduplicateDataSet(ctx context.Context) error {
	d.DeduplicateDataSetInvocations++

	d.DeduplicateDataSetInputs = append(d.DeduplicateDataSetInputs, ctx)

	gomega.Expect(d.DeduplicateDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.DeduplicateDataSetOutputs[0]
	d.DeduplicateDataSetOutputs = d.DeduplicateDataSetOutputs[1:]
	return output
}

func (d *Deduplicator) DeleteDataSet(ctx context.Context) error {
	d.DeleteDataSetInvocations++

	d.DeleteDataSetInputs = append(d.DeleteDataSetInputs, ctx)

	gomega.Expect(d.DeleteDataSetOutputs).ToNot(gomega.BeEmpty())

	output := d.DeleteDataSetOutputs[0]
	d.DeleteDataSetOutputs = d.DeleteDataSetOutputs[1:]
	return output
}

func (d *Deduplicator) Expectations() {
	d.Mock.Expectations()
	gomega.Expect(d.NameOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.VersionOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.RegisterDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.AddDataSetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DeduplicateDataSetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DeleteDataSetOutputs).To(gomega.BeEmpty())
}
