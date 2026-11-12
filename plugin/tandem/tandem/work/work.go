package work

import (
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

type DataClient any

type DataDeduplicatorFactory any

type DataRawClient any

type DataSetClient any

type DataSourceClient any

type CloudDriversClient any

type ProcessorDependencies struct {
	workBase.Dependencies
	DataClient              DataClient
	DataDeduplicatorFactory DataDeduplicatorFactory
	DataRawClient           DataRawClient
	DataSetClient           DataSetClient
	DataSourceClient        DataSourceClient
	CloudDriversClient      CloudDriversClient
}

func NewProcessorFactories(processorDependencies ProcessorDependencies) ([]work.ProcessorFactory, error) {
	return nil, nil
}
