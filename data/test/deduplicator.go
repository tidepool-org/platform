package test

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/test"
)

type Deduplicator struct {
	*test.Mock
	NameInvocations               int
	NameOutputs                   []string
	VersionInvocations            int
	VersionOutputs                []string
	RegisterDatasetInvocations    int
	RegisterDatasetOutputs        []error
	AddDatasetDataInvocations     int
	AddDatasetDataInputs          [][]data.Datum
	AddDatasetDataOutputs         []error
	DeduplicateDatasetInvocations int
	DeduplicateDatasetOutputs     []error
	DeleteDatasetInvocations      int
	DeleteDatasetOutputs          []error
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		Mock: test.NewMock(),
	}
}

func (d *Deduplicator) Name() string {
	d.NameInvocations++

	if len(d.NameOutputs) == 0 {
		panic("Unexpected invocation of Name on Deduplicator")
	}

	output := d.NameOutputs[0]
	d.NameOutputs = d.NameOutputs[1:]
	return output
}

func (d *Deduplicator) Version() string {
	d.VersionInvocations++

	if len(d.VersionOutputs) == 0 {
		panic("Unexpected invocation of Version on Deduplicator")
	}

	output := d.VersionOutputs[0]
	d.VersionOutputs = d.VersionOutputs[1:]
	return output
}

func (d *Deduplicator) RegisterDataset() error {
	d.RegisterDatasetInvocations++

	if len(d.RegisterDatasetOutputs) == 0 {
		panic("Unexpected invocation of RegisterDataset on Deduplicator")
	}

	output := d.RegisterDatasetOutputs[0]
	d.RegisterDatasetOutputs = d.RegisterDatasetOutputs[1:]
	return output
}

func (d *Deduplicator) AddDatasetData(datasetData []data.Datum) error {
	d.AddDatasetDataInvocations++

	d.AddDatasetDataInputs = append(d.AddDatasetDataInputs, datasetData)

	if len(d.AddDatasetDataOutputs) == 0 {
		panic("Unexpected invocation of AddDatasetData on Deduplicator")
	}

	output := d.AddDatasetDataOutputs[0]
	d.AddDatasetDataOutputs = d.AddDatasetDataOutputs[1:]
	return output
}

func (d *Deduplicator) DeduplicateDataset() error {
	d.DeduplicateDatasetInvocations++

	if len(d.DeduplicateDatasetOutputs) == 0 {
		panic("Unexpected invocation of DeduplicateDataset on Deduplicator")
	}

	output := d.DeduplicateDatasetOutputs[0]
	d.DeduplicateDatasetOutputs = d.DeduplicateDatasetOutputs[1:]
	return output
}

func (d *Deduplicator) DeleteDataset() error {
	d.DeleteDatasetInvocations++

	if len(d.DeleteDatasetOutputs) == 0 {
		panic("Unexpected invocation of DeleteDataset on Deduplicator")
	}

	output := d.DeleteDatasetOutputs[0]
	d.DeleteDatasetOutputs = d.DeleteDatasetOutputs[1:]
	return output
}

func (d *Deduplicator) UnusedOutputsCount() int {
	return len(d.NameOutputs) +
		len(d.VersionOutputs) +
		len(d.RegisterDatasetOutputs) +
		len(d.AddDatasetDataOutputs) +
		len(d.DeduplicateDatasetOutputs) +
		len(d.DeleteDatasetOutputs)
}
