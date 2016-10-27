package test

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
)

type CanDeduplicateDatasetOutput struct {
	Can   bool
	Error error
}

type NewDeduplicatorInput struct {
	Logger           log.Logger
	DataStoreSession store.Session
	Dataset          *upload.Upload
}

type NewDeduplicatorOutput struct {
	Deduplicator data.Deduplicator
	Error        error
}

type Factory struct {
	ID                               string
	CanDeduplicateDatasetInvocations int
	CanDeduplicateDatasetInputs      []*upload.Upload
	CanDeduplicateDatasetOutputs     []CanDeduplicateDatasetOutput
	NewDeduplicatoInvocations        int
	NewDeduplicatorInputs            []NewDeduplicatorInput
	NewDeduplicatorOutputs           []NewDeduplicatorOutput
}

func NewFactory() *Factory {
	return &Factory{
		ID: app.NewID(),
	}
}

func (f *Factory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	f.CanDeduplicateDatasetInvocations++

	f.CanDeduplicateDatasetInputs = append(f.CanDeduplicateDatasetInputs, dataset)

	if len(f.CanDeduplicateDatasetOutputs) == 0 {
		panic("Unexpected invocation of CanDeduplicateDataset on Factory")
	}

	output := f.CanDeduplicateDatasetOutputs[0]
	f.CanDeduplicateDatasetOutputs = f.CanDeduplicateDatasetOutputs[1:]
	return output.Can, output.Error
}

func (f *Factory) NewDeduplicator(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (data.Deduplicator, error) {
	f.NewDeduplicatoInvocations++

	f.NewDeduplicatorInputs = append(f.NewDeduplicatorInputs, NewDeduplicatorInput{logger, dataStoreSession, dataset})

	if len(f.NewDeduplicatorOutputs) == 0 {
		panic("Unexpected invocation of NewDeduplicator on Factory")
	}

	output := f.NewDeduplicatorOutputs[0]
	f.NewDeduplicatorOutputs = f.NewDeduplicatorOutputs[1:]
	return output.Deduplicator, output.Error
}

func (f *Factory) UnusedOutputsCount() int {
	return len(f.CanDeduplicateDatasetOutputs) +
		len(f.NewDeduplicatorOutputs)
}
