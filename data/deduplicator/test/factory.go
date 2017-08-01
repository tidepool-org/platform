package test

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
)

type CanDeduplicateDatasetOutput struct {
	Can   bool
	Error error
}

type NewDeduplicatorForDatasetInput struct {
	Logger           log.Logger
	DataStoreSession store.Session
	Dataset          *upload.Upload
}

type NewDeduplicatorForDatasetOutput struct {
	Deduplicator data.Deduplicator
	Error        error
}

type IsRegisteredWithDatasetOutput struct {
	Is    bool
	Error error
}

type NewRegisteredDeduplicatorForDatasetInput struct {
	Logger           log.Logger
	DataStoreSession store.Session
	Dataset          *upload.Upload
}

type NewRegisteredDeduplicatorForDatasetOutput struct {
	Deduplicator data.Deduplicator
	Error        error
}

type Factory struct {
	ID                                             string
	CanDeduplicateDatasetInvocations               int
	CanDeduplicateDatasetInputs                    []*upload.Upload
	CanDeduplicateDatasetOutputs                   []CanDeduplicateDatasetOutput
	NewDeduplicatorForDatasetInvocations           int
	NewDeduplicatorForDatasetInputs                []NewDeduplicatorForDatasetInput
	NewDeduplicatorForDatasetOutputs               []NewDeduplicatorForDatasetOutput
	IsRegisteredWithDatasetInvocations             int
	IsRegisteredWithDatasetInputs                  []*upload.Upload
	IsRegisteredWithDatasetOutputs                 []IsRegisteredWithDatasetOutput
	NewRegisteredDeduplicatorForDatasetInvocations int
	NewRegisteredDeduplicatorForDatasetInputs      []NewRegisteredDeduplicatorForDatasetInput
	NewRegisteredDeduplicatorForDatasetOutputs     []NewRegisteredDeduplicatorForDatasetOutput
}

func NewFactory() *Factory {
	return &Factory{
		ID: id.New(),
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

func (f *Factory) NewDeduplicatorForDataset(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (data.Deduplicator, error) {
	f.NewDeduplicatorForDatasetInvocations++

	f.NewDeduplicatorForDatasetInputs = append(f.NewDeduplicatorForDatasetInputs, NewDeduplicatorForDatasetInput{logger, dataStoreSession, dataset})

	if len(f.NewDeduplicatorForDatasetOutputs) == 0 {
		panic("Unexpected invocation of NewDeduplicatorForDataset on Factory")
	}

	output := f.NewDeduplicatorForDatasetOutputs[0]
	f.NewDeduplicatorForDatasetOutputs = f.NewDeduplicatorForDatasetOutputs[1:]
	return output.Deduplicator, output.Error
}

func (f *Factory) IsRegisteredWithDataset(dataset *upload.Upload) (bool, error) {
	f.IsRegisteredWithDatasetInvocations++

	f.IsRegisteredWithDatasetInputs = append(f.IsRegisteredWithDatasetInputs, dataset)

	if len(f.IsRegisteredWithDatasetOutputs) == 0 {
		panic("Unexpected invocation of IsRegisteredWithDataset on Factory")
	}

	output := f.IsRegisteredWithDatasetOutputs[0]
	f.IsRegisteredWithDatasetOutputs = f.IsRegisteredWithDatasetOutputs[1:]
	return output.Is, output.Error
}

func (f *Factory) NewRegisteredDeduplicatorForDataset(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (data.Deduplicator, error) {
	f.NewRegisteredDeduplicatorForDatasetInvocations++

	f.NewRegisteredDeduplicatorForDatasetInputs = append(f.NewRegisteredDeduplicatorForDatasetInputs, NewRegisteredDeduplicatorForDatasetInput{logger, dataStoreSession, dataset})

	if len(f.NewRegisteredDeduplicatorForDatasetOutputs) == 0 {
		panic("Unexpected invocation of NewRegisteredDeduplicatorForDataset on Factory")
	}

	output := f.NewRegisteredDeduplicatorForDatasetOutputs[0]
	f.NewRegisteredDeduplicatorForDatasetOutputs = f.NewRegisteredDeduplicatorForDatasetOutputs[1:]
	return output.Deduplicator, output.Error
}

func (f *Factory) UnusedOutputsCount() int {
	return len(f.CanDeduplicateDatasetOutputs) +
		len(f.NewDeduplicatorForDatasetOutputs) +
		len(f.IsRegisteredWithDatasetOutputs) +
		len(f.NewRegisteredDeduplicatorForDatasetOutputs)
}
