package provider

import (
	"context"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/data"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/dexcom/fetch"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

const ProviderName = "dexcom"

type Provider struct {
	*oauthProvider.Provider
	dataClient dataClient.Client
	taskClient task.Client
}

func New(configReporter config.Reporter, dataClient dataClient.Client, taskClient task.Client) (*Provider, error) {
	if configReporter == nil {
		return nil, errors.New("config reporter is missing")
	}
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}
	if taskClient == nil {
		return nil, errors.New("task client is missing")
	}

	prvdr, err := oauthProvider.NewProvider(ProviderName, configReporter.WithScopes(ProviderName))
	if err != nil {
		return nil, err
	}

	return &Provider{
		Provider:   prvdr,
		dataClient: dataClient,
		taskClient: taskClient,
	}, nil
}

func (p *Provider) OnCreate(ctx context.Context, userID string, providerSessionID string) error {
	if userID == "" {
		return errors.New("user id is missing")
	}
	if providerSessionID == "" {
		return errors.New("provider session id is missing")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "type": p.Type(), "name": p.Name()})

	filter := data.NewDataSourceFilter()
	filter.ProviderType = pointer.FromString(p.Type())
	filter.ProviderName = pointer.FromString(p.Name())
	dataSources, err := p.dataClient.ListUserDataSources(ctx, userID, filter, nil)
	if err != nil {
		return errors.Wrap(err, "unable to fetch data sources")
	}

	var dataSource *data.DataSource
	if dataSourcesCount := len(dataSources); dataSourcesCount > 0 {
		if dataSourcesCount > 1 {
			logger.WithField("dataSourcesCount", dataSourcesCount).Warn("unexpected number of data sources found")
		}

		dataSource = dataSources[0]
		if dataSource.State != data.DataSourceStateDisconnected {
			logger.WithFields(log.Fields{"dataSourceId": dataSource.ID, "dataSourceState": dataSource.State}).Warn("data source in unexpected state")
		}

		dataSourceUpdate := data.NewDataSourceUpdate()
		dataSourceUpdate.State = pointer.FromString(data.DataSourceStateConnected)

		dataSource, err = p.dataClient.UpdateDataSource(ctx, dataSource.ID, dataSourceUpdate)
		if err != nil {
			return errors.Wrap(err, "unable to update data source")
		}
	} else {
		dataSourceCreate := data.NewDataSourceCreate()
		dataSourceCreate.ProviderType = p.Type()
		dataSourceCreate.ProviderName = p.Name()
		dataSourceCreate.ProviderSessionID = providerSessionID
		dataSourceCreate.State = data.DataSourceStateConnected

		dataSource, err = p.dataClient.CreateUserDataSource(ctx, userID, dataSourceCreate)
		if err != nil {
			return errors.Wrap(err, "unable to create data source")
		}
	}

	taskCreate, err := fetch.NewTaskCreate(providerSessionID, dataSource.ID)
	if err != nil {
		return errors.Wrap(err, "unable to create task create")
	}

	_, err = p.taskClient.CreateTask(ctx, taskCreate)
	if err != nil {
		p.dataClient.DeleteDataSource(ctx, dataSource.ID)
		return errors.Wrap(err, "unable to create task")
	}

	return nil
}

func (p *Provider) OnDelete(ctx context.Context, userID string, providerSessionID string) error {
	if userID == "" {
		return errors.New("user id is missing")
	}
	if providerSessionID == "" {
		return errors.New("provider session id is missing")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "providerSessionId": providerSessionID})

	taskFilter := task.NewTaskFilter()
	taskFilter.Name = pointer.FromString(fetch.TaskName(providerSessionID))
	tasks, err := p.taskClient.ListTasks(ctx, taskFilter, nil)
	if err != nil {
		logger.WithError(err).Error("unable to list tasks after deleting provider session")
		return nil
	}

	for _, task := range tasks {
		if err = p.taskClient.DeleteTask(ctx, task.ID); err != nil {
			logger.WithError(err).WithField("taskId", task.ID).Error("unable to delete task after deleting provider session")
		}
		if dataSourceID, ok := task.Data["dataSourceId"].(string); ok && dataSourceID != "" {
			dataSourceUpdate := data.NewDataSourceUpdate()
			dataSourceUpdate.State = pointer.FromString(data.DataSourceStateDisconnected)
			_, err = p.dataClient.UpdateDataSource(ctx, dataSourceID, dataSourceUpdate)
			if err != nil {
				logger.WithError(err).WithField("dataSourceId", dataSourceID).Error("unable to update data source after deleting provider session")
			}
		}
	}
	return nil
}
