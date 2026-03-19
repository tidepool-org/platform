package processors

import (
	"context"

	providerSession "github.com/tidepool-org/platform/auth/providersession"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataSet "github.com/tidepool-org/platform/data/set"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura"
	ouraDataWorkEvent "github.com/tidepool-org/platform/oura/data/work/event"
	ouraDataWorkHistoric "github.com/tidepool-org/platform/oura/data/work/historic"
	ouraDataWorkSetup "github.com/tidepool-org/platform/oura/data/work/setup"
	ouraUsersWorkRevoke "github.com/tidepool-org/platform/oura/users/work/revoke"
	ouraWebhookWorkSubscribe "github.com/tidepool-org/platform/oura/webhook/work/subscribe"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

type (
	ProviderSessionClient = providerSession.Client
	DataSourceClient      = dataSource.Client
	DataRawClient         = dataRaw.Client
	DataSetClient         = dataSet.Client
	OuraClient            = oura.Client
)

type Dependencies struct {
	workBase.Dependencies
	ProviderSessionClient
	DataSourceClient
	DataRawClient
	DataSetClient
	OuraClient
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.ProviderSessionClient == nil {
		return errors.New("provider session client is missing")
	}
	if d.DataSourceClient == nil {
		return errors.New("data source client is missing")
	}
	if d.DataRawClient == nil {
		return errors.New("data raw client is missing")
	}
	if d.DataSetClient == nil {
		return errors.New("data set client is missing")
	}
	if d.OuraClient == nil {
		return errors.New("oura client is missing")
	}
	return nil
}

func NewProcessorFactories(dependencies Dependencies) ([]work.ProcessorFactory, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	var processorFactories []work.ProcessorFactory

	if processorFactory, err := ouraDataWorkEvent.NewProcessorFactory(ouraDataWorkEvent.Dependencies{
		Dependencies:          dependencies.Dependencies,
		ProviderSessionClient: dependencies.ProviderSessionClient,
		DataSourceClient:      dependencies.DataSourceClient,
		OuraClient:            dependencies.OuraClient,
	}); err != nil {
		return nil, err
	} else {
		processorFactories = append(processorFactories, processorFactory)
	}

	if processorFactory, err := ouraDataWorkHistoric.NewProcessorFactory(ouraDataWorkHistoric.Dependencies{
		Dependencies:          dependencies.Dependencies,
		ProviderSessionClient: dependencies.ProviderSessionClient,
		DataSourceClient:      dependencies.DataSourceClient,
		OuraClient:            dependencies.OuraClient,
	}); err != nil {
		return nil, err
	} else {
		processorFactories = append(processorFactories, processorFactory)
	}

	if processorFactory, err := ouraDataWorkSetup.NewProcessorFactory(ouraDataWorkSetup.Dependencies{
		Dependencies:          dependencies.Dependencies,
		ProviderSessionClient: dependencies.ProviderSessionClient,
		DataSourceClient:      dependencies.DataSourceClient,
		OuraClient:            dependencies.OuraClient,
	}); err != nil {
		return nil, err
	} else {
		processorFactories = append(processorFactories, processorFactory)
	}

	if processorFactory, err := ouraWebhookWorkSubscribe.NewProcessorFactory(ouraWebhookWorkSubscribe.Dependencies{
		Dependencies: dependencies.Dependencies,
		OuraClient:   dependencies.OuraClient,
	}); err != nil {
		return nil, err
	} else {
		processorFactories = append(processorFactories, processorFactory)
	}

	if processorFactory, err := ouraUsersWorkRevoke.NewProcessorFactory(ouraUsersWorkRevoke.Dependencies{
		Dependencies: dependencies.Dependencies,
		OuraClient:   dependencies.OuraClient,
	}); err != nil {
		return nil, err
	} else {
		processorFactories = append(processorFactories, processorFactory)
	}

	return processorFactories, nil
}

func EnsureWork(ctx context.Context, workClient work.Client) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if workClient == nil {
		return errors.New("work client is missing")
	}

	if workCreate, err := ouraWebhookWorkSubscribe.NewWorkCreate(); err != nil {
		return errors.Wrap(err, "unable to create webhook subscribe work create")
	} else if _, err = workClient.Create(ctx, workCreate); err != nil {
		return errors.Wrap(err, "unable to create webhook subscribe work")
	}

	log.LoggerFromContext(ctx).Debug("created webhook subscribe work")
	return nil
}
