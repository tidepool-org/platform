package work

import (
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

type DataDeduplicatorFactory any

type DataRawClient any

type DataSetClient any

type DataSourceClient any

type SummaryClient any

type ProviderSessionClient any

type AbbottClient any

type WorkClient any

type ProcessorDependencies struct {
	workBase.Dependencies
	DataDeduplicatorFactory DataDeduplicatorFactory
	DataRawClient           DataRawClient
	DataSetClient           DataSetClient
	DataSourceClient        DataSourceClient
	SummaryClient           SummaryClient
	ProviderSessionClient   ProviderSessionClient
	AbbottClient            AbbottClient
}

func NewProcessorFactories(processorDependencies ProcessorDependencies) ([]work.ProcessorFactory, error) {
	return nil, nil
}
