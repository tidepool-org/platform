package provider

import (
	"context"

	"github.com/tidepool-org/platform/config"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/dexcom/fetch"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

const ProviderName = "twiist"

type Provider struct {
	*oauthProvider.Provider
	dataSourceClient dataSource.Client
	taskClient       task.Client
}

func New(configReporter config.Reporter, dataSourceClient dataSource.Client, taskClient task.Client) (*Provider, error) {
	if configReporter == nil {
		return nil, errors.New("config reporter is missing")
	}
	if dataSourceClient == nil {
		return nil, errors.New("data source client is missing")
	}
	if taskClient == nil {
		return nil, errors.New("task client is missing")
	}

	prvdr, err := oauthProvider.NewProvider(ProviderName, configReporter.WithScopes(ProviderName))
	if err != nil {
		return nil, err
	}

	return &Provider{
		Provider:         prvdr,
		dataSourceClient: dataSourceClient,
		taskClient:       taskClient,
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

	filter := dataSource.NewFilter()
	filter.ProviderType = pointer.FromStringArray([]string{p.Type()})
	filter.ProviderName = pointer.FromStringArray([]string{p.Name()})
	sources, err := p.dataSourceClient.List(ctx, userID, filter, nil)
	if err != nil {
		return errors.Wrap(err, "unable to fetch data sources")
	}

	var source *dataSource.Source
	if count := len(sources); count > 0 {
		if count > 1 {
			logger.WithField("count", count).Warn("unexpected number of data sources found")
		}

		source = sources[0]
		if *source.State != dataSource.StateDisconnected {
			logger.WithFields(log.Fields{"id": source.ID, "state": source.State}).Warn("data source in unexpected state")
		}

		update := dataSource.NewUpdate()
		update.ProviderSessionID = pointer.FromString(providerSessionID)
		update.State = pointer.FromString(dataSource.StateConnected)

		source, err = p.dataSourceClient.Update(ctx, *source.ID, nil, update)
		if err != nil {
			return errors.Wrap(err, "unable to update data source")
		}
	} else {
		create := dataSource.NewCreate()
		create.ProviderType = pointer.FromString(p.Type())
		create.ProviderName = pointer.FromString(p.Name())
		create.ProviderSessionID = pointer.FromString(providerSessionID)
		create.State = pointer.FromString(dataSource.StateConnected)

		source, err = p.dataSourceClient.Create(ctx, userID, create)
		if err != nil {
			return errors.Wrap(err, "unable to create data source")
		}
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
			update := dataSource.NewUpdate()
			update.State = pointer.FromString(dataSource.StateDisconnected)
			_, err = p.dataSourceClient.Update(ctx, dataSourceID, nil, update)
			if err != nil {
				logger.WithError(err).WithField("dataSourceId", dataSourceID).Error("unable to update data source after deleting provider session")
			}
		}
	}
	return nil
}
