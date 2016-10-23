package test

import "github.com/tidepool-org/platform/data"

type Deduplicator struct {
	InitializeDatasetInvocations int
	InitializeDatasetOutputs     []error
	AddDataToDatasetInvocations  int
	AddDataToDatasetInputs       [][]data.Datum
	AddDataToDatasetOutputs      []error
	FinalizeDatasetInvocations   int
	FinalizeDatasetOutputs       []error
}

func (d *Deduplicator) InitializeDataset() error {
	d.InitializeDatasetInvocations++

	if len(d.InitializeDatasetOutputs) == 0 {
		panic("Unexpected invocation of InitializeDataset on Deduplicator")
	}

	output := d.InitializeDatasetOutputs[0]
	d.InitializeDatasetOutputs = d.InitializeDatasetOutputs[1:]
	return output
}

func (d *Deduplicator) AddDataToDataset(datasetData []data.Datum) error {
	d.AddDataToDatasetInvocations++

	d.AddDataToDatasetInputs = append(d.AddDataToDatasetInputs, datasetData)

	if len(d.AddDataToDatasetOutputs) == 0 {
		panic("Unexpected invocation of AddDataToDataset on Deduplicator")
	}

	output := d.AddDataToDatasetOutputs[0]
	d.AddDataToDatasetOutputs = d.AddDataToDatasetOutputs[1:]
	return output
}

func (d *Deduplicator) FinalizeDataset() error {
	d.FinalizeDatasetInvocations++

	if len(d.FinalizeDatasetOutputs) == 0 {
		panic("Unexpected invocation of FinalizeDataset on Deduplicator")
	}

	output := d.FinalizeDatasetOutputs[0]
	d.FinalizeDatasetOutputs = d.FinalizeDatasetOutputs[1:]
	return output
}

func (d *Deduplicator) UnusedOutputsCount() int {
	return len(d.InitializeDatasetOutputs) +
		len(d.AddDataToDatasetOutputs) +
		len(d.FinalizeDatasetOutputs)
}
