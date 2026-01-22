package work

import (
	providerSession "github.com/tidepool-org/platform/auth/providersession"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataSet "github.com/tidepool-org/platform/data/set"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataWork "github.com/tidepool-org/platform/data/work"
	ouraWork "github.com/tidepool-org/platform/oura/work"
	ouraWorkDataEvent "github.com/tidepool-org/platform/oura/work/data/event"
	ouraWorkDataHistoric "github.com/tidepool-org/platform/oura/work/data/historic"
	ouraWorkDataSetup "github.com/tidepool-org/platform/oura/work/data/setup"
	ouraWorkSubscribe "github.com/tidepool-org/platform/oura/work/subscribe"
	ouraWorkUsersRevoke "github.com/tidepool-org/platform/oura/work/users/revoke"
	"github.com/tidepool-org/platform/work"
)

type Dependencies struct {
	ProviderSessionClient providerSession.Client
	DataSourceClient      dataSource.Client
	DataRawClient         dataRaw.Client
	DataSetClient         dataSet.Client
	WorkClient            work.Client
	Client                ouraWork.Client
}

func NewProcessorFactories(dependencies Dependencies) ([]work.ProcessorFactory, error) {
	var processorFactories []work.ProcessorFactory

	if processorFactory, err := ouraWorkDataEvent.NewProcessorFactory(ouraWorkDataEvent.Dependencies{
		DataDependencies: dataWork.Dependencies{
			ProviderSessionClient: dependencies.ProviderSessionClient,
			DataSourceClient:      dependencies.DataSourceClient,
			DataRawClient:         dependencies.DataRawClient,
			DataSetClient:         dependencies.DataSetClient,
		},
		Client: dependencies.Client,
	}); err != nil {
		return nil, err
	} else {
		processorFactories = append(processorFactories, processorFactory)
	}

	if processorFactory, err := ouraWorkDataHistoric.NewProcessorFactory(ouraWorkDataHistoric.Dependencies{
		DataDependencies: dataWork.Dependencies{
			ProviderSessionClient: dependencies.ProviderSessionClient,
			DataSourceClient:      dependencies.DataSourceClient,
			DataRawClient:         dependencies.DataRawClient,
			DataSetClient:         dependencies.DataSetClient,
		},
		Client: dependencies.Client,
	}); err != nil {
		return nil, err
	} else {
		processorFactories = append(processorFactories, processorFactory)
	}

	if processorFactory, err := ouraWorkDataSetup.NewProcessorFactory(ouraWorkDataSetup.Dependencies{
		ProviderSessionClient: dependencies.ProviderSessionClient,
		DataSourceClient:      dependencies.DataSourceClient,
		WorkClient:            dependencies.WorkClient,
		Client:                dependencies.Client,
	}); err != nil {
		return nil, err
	} else {
		processorFactories = append(processorFactories, processorFactory)
	}

	if processorFactory, err := ouraWorkSubscribe.NewProcessorFactory(ouraWorkSubscribe.Dependencies{
		Client: dependencies.Client,
	}); err != nil {
		return nil, err
	} else {
		processorFactories = append(processorFactories, processorFactory)
	}

	if processorFactory, err := ouraWorkUsersRevoke.NewProcessorFactory(ouraWorkUsersRevoke.Dependencies{
		Client: dependencies.Client,
	}); err != nil {
		return nil, err
	} else {
		processorFactories = append(processorFactories, processorFactory)
	}

	return processorFactories, nil
}
