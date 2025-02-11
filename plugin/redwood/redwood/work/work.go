package work

import (
	"github.com/tidepool-org/platform/work"
)

type DataClient any

type DataRawClient any

type DataSetClient any

type DataSourceClient any

type ProviderSessionClient any

type RedwoodClient any

type WorkClient any

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
