package work

import (
	"github.com/tidepool-org/platform/work"
)

type DataClient interface{}

type DataRawClient interface{}

type DataSetClient interface{}

type DataSourceClient interface{}

type ProviderSessionClient interface{}

type RedwoodClient interface{}

type WorkClient interface{}

type ProcessorDependencies struct {
	DataClient            DataClient
	DataRawClient         DataRawClient
	DataSetClient         DataSetClient
	DataSourceClient      DataSourceClient
	ProviderSessionClient ProviderSessionClient
	RedwoodClient         RedwoodClient
	WorkClient            WorkClient
}

func NewProcessors(processorDependencies ProcessorDependencies) ([]work.Processor, error) {
	return nil, nil
}
