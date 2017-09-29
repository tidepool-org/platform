package test

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/test"
)

type AddDatasetDataInput struct {
	Context     context.Context
	DatasetData []data.Datum
}

type Deduplicator struct {
	*test.Mock
	NameInvocations               int
	NameOutputs                   []string
	VersionInvocations            int
	VersionOutputs                []string
	RegisterDatasetInvocations    int
	RegisterDatasetInputs         []context.Context
	RegisterDatasetOutputs        []error
	AddDatasetDataInvocations     int
	AddDatasetDataInputs          []AddDatasetDataInput
	AddDatasetDataOutputs         []error
	DeduplicateDatasetInvocations int
	DeduplicateDatasetInputs      []context.Context
	DeduplicateDatasetOutputs     []error
	DeleteDatasetInvocations      int
	DeleteDatasetInputs           []context.Context
	DeleteDatasetOutputs          []error
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

func (d *Deduplicator) RegisterDataset(ctx context.Context) error {
	d.RegisterDatasetInvocations++

	d.RegisterDatasetInputs = append(d.RegisterDatasetInputs, ctx)

	gomega.Expect(d.RegisterDatasetOutputs).ToNot(gomega.BeEmpty())

	output := d.RegisterDatasetOutputs[0]
	d.RegisterDatasetOutputs = d.RegisterDatasetOutputs[1:]
	return output
}

func (d *Deduplicator) AddDatasetData(ctx context.Context, datasetData []data.Datum) error {
	d.AddDatasetDataInvocations++

	d.AddDatasetDataInputs = append(d.AddDatasetDataInputs, AddDatasetDataInput{Context: ctx, DatasetData: datasetData})

	gomega.Expect(d.AddDatasetDataOutputs).ToNot(gomega.BeEmpty())

	output := d.AddDatasetDataOutputs[0]
	d.AddDatasetDataOutputs = d.AddDatasetDataOutputs[1:]
	return output
}

func (d *Deduplicator) DeduplicateDataset(ctx context.Context) error {
	d.DeduplicateDatasetInvocations++

	d.DeduplicateDatasetInputs = append(d.DeduplicateDatasetInputs, ctx)

	gomega.Expect(d.DeduplicateDatasetOutputs).ToNot(gomega.BeEmpty())

	output := d.DeduplicateDatasetOutputs[0]
	d.DeduplicateDatasetOutputs = d.DeduplicateDatasetOutputs[1:]
	return output
}

func (d *Deduplicator) DeleteDataset(ctx context.Context) error {
	d.DeleteDatasetInvocations++

	d.DeleteDatasetInputs = append(d.DeleteDatasetInputs, ctx)

	gomega.Expect(d.DeleteDatasetOutputs).ToNot(gomega.BeEmpty())

	output := d.DeleteDatasetOutputs[0]
	d.DeleteDatasetOutputs = d.DeleteDatasetOutputs[1:]
	return output
}

func (d *Deduplicator) Expectations() {
	d.Mock.Expectations()
	gomega.Expect(d.NameOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.VersionOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.RegisterDatasetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.AddDatasetDataOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DeduplicateDatasetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.DeleteDatasetOutputs).To(gomega.BeEmpty())
}
