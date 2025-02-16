package work

import (
	"github.com/tidepool-org/platform/work"
)

type DataDeduplicatorFactory any

type DataClient any

type DataRawClient any

type DataSetClient any

type DataSourceClient any

type ProviderSessionClient any

type AbbottClient any

type WorkClient any

type ProcessorDependencies struct {
	DataDeduplicatorFactory DataDeduplicatorFactory
	DataClient              DataClient
	DataRawClient           DataRawClient
	DataSetClient           DataSetClient
	DataSourceClient        DataSourceClient
	ProviderSessionClient   ProviderSessionClient
	AbbottClient            AbbottClient
	WorkClient              WorkClient
}

func NewProcessors(processorDependencies ProcessorDependencies) ([]work.Processor, error) {
	return nil, nil
}
