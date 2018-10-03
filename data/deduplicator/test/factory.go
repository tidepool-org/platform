package test

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/test"
)

type CanDeduplicateDataSetOutput struct {
	Can   bool
	Error error
}

type NewDeduplicatorForDataSetInput struct {
	Logger      log.Logger
	DataSession storeDEPRECATED.DataSession
	DataSet     *upload.Upload
}

type NewDeduplicatorForDataSetOutput struct {
	Deduplicator data.Deduplicator
	Error        error
}

type IsRegisteredWithDataSetOutput struct {
	Is    bool
	Error error
}

type NewRegisteredDeduplicatorForDataSetInput struct {
	Logger      log.Logger
	DataSession storeDEPRECATED.DataSession
	DataSet     *upload.Upload
}

type NewRegisteredDeduplicatorForDataSetOutput struct {
	Deduplicator data.Deduplicator
	Error        error
}

type Factory struct {
	*test.Mock
	CanDeduplicateDataSetInvocations               int
	CanDeduplicateDataSetInputs                    []*upload.Upload
	CanDeduplicateDataSetOutputs                   []CanDeduplicateDataSetOutput
	NewDeduplicatorForDataSetInvocations           int
	NewDeduplicatorForDataSetInputs                []NewDeduplicatorForDataSetInput
	NewDeduplicatorForDataSetOutputs               []NewDeduplicatorForDataSetOutput
	IsRegisteredWithDataSetInvocations             int
	IsRegisteredWithDataSetInputs                  []*upload.Upload
	IsRegisteredWithDataSetOutputs                 []IsRegisteredWithDataSetOutput
	NewRegisteredDeduplicatorForDataSetInvocations int
	NewRegisteredDeduplicatorForDataSetInputs      []NewRegisteredDeduplicatorForDataSetInput
	NewRegisteredDeduplicatorForDataSetOutputs     []NewRegisteredDeduplicatorForDataSetOutput
}

func NewFactory() *Factory {
	return &Factory{
		Mock: test.NewMock(),
	}
}

func (f *Factory) CanDeduplicateDataSet(dataSet *upload.Upload) (bool, error) {
	f.CanDeduplicateDataSetInvocations++

	f.CanDeduplicateDataSetInputs = append(f.CanDeduplicateDataSetInputs, dataSet)

	if len(f.CanDeduplicateDataSetOutputs) == 0 {
		panic("Unexpected invocation of CanDeduplicateDataSet on Factory")
	}

	output := f.CanDeduplicateDataSetOutputs[0]
	f.CanDeduplicateDataSetOutputs = f.CanDeduplicateDataSetOutputs[1:]
	return output.Can, output.Error
}

func (f *Factory) NewDeduplicatorForDataSet(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataSet *upload.Upload) (data.Deduplicator, error) {
	f.NewDeduplicatorForDataSetInvocations++

	f.NewDeduplicatorForDataSetInputs = append(f.NewDeduplicatorForDataSetInputs, NewDeduplicatorForDataSetInput{logger, dataSession, dataSet})

	if len(f.NewDeduplicatorForDataSetOutputs) == 0 {
		panic("Unexpected invocation of NewDeduplicatorForDataSet on Factory")
	}

	output := f.NewDeduplicatorForDataSetOutputs[0]
	f.NewDeduplicatorForDataSetOutputs = f.NewDeduplicatorForDataSetOutputs[1:]
	return output.Deduplicator, output.Error
}

func (f *Factory) IsRegisteredWithDataSet(dataSet *upload.Upload) (bool, error) {
	f.IsRegisteredWithDataSetInvocations++

	f.IsRegisteredWithDataSetInputs = append(f.IsRegisteredWithDataSetInputs, dataSet)

	if len(f.IsRegisteredWithDataSetOutputs) == 0 {
		panic("Unexpected invocation of IsRegisteredWithDataSet on Factory")
	}

	output := f.IsRegisteredWithDataSetOutputs[0]
	f.IsRegisteredWithDataSetOutputs = f.IsRegisteredWithDataSetOutputs[1:]
	return output.Is, output.Error
}

func (f *Factory) NewRegisteredDeduplicatorForDataSet(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataSet *upload.Upload) (data.Deduplicator, error) {
	f.NewRegisteredDeduplicatorForDataSetInvocations++

	f.NewRegisteredDeduplicatorForDataSetInputs = append(f.NewRegisteredDeduplicatorForDataSetInputs, NewRegisteredDeduplicatorForDataSetInput{logger, dataSession, dataSet})

	if len(f.NewRegisteredDeduplicatorForDataSetOutputs) == 0 {
		panic("Unexpected invocation of NewRegisteredDeduplicatorForDataSet on Factory")
	}

	output := f.NewRegisteredDeduplicatorForDataSetOutputs[0]
	f.NewRegisteredDeduplicatorForDataSetOutputs = f.NewRegisteredDeduplicatorForDataSetOutputs[1:]
	return output.Deduplicator, output.Error
}

func (f *Factory) UnusedOutputsCount() int {
	return len(f.CanDeduplicateDataSetOutputs) +
		len(f.NewDeduplicatorForDataSetOutputs) +
		len(f.IsRegisteredWithDataSetOutputs) +
		len(f.NewRegisteredDeduplicatorForDataSetOutputs)
}
