package provider

import (
	"context"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/config"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/dexcom"
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

func (p *Provider) OnCreate(ctx context.Context, userID string, providerSession *auth.ProviderSession) (*auth.ProviderSessionUpdate, error) {
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if providerSession == nil {
		return nil, errors.New("provider session is missing")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "type": p.Type(), "name": p.Name()})

	filter := dataSource.NewFilter()
	filter.ProviderType = pointer.FromStringArray([]string{p.Type()})
	filter.ProviderName = pointer.FromStringArray([]string{p.Name()})
	sources, err := p.dataSourceClient.List(ctx, userID, filter, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch data sources")
	}

	var source *dataSource.Source
	if count := len(sources); count > 0 {
		if count > 1 {
			logger.WithField("count", count).Warn("unexpected number of data sources found")
		}

		for _, source := range sources {
			if *source.State != dataSource.StateDisconnected {
				logger.WithFields(log.Fields{"id": source.ID, "state": source.State}).Warn("data source in unexpected state")

				update := dataSource.NewUpdate()
				update.State = pointer.FromString(dataSource.StateDisconnected)

				_, err = p.dataSourceClient.Update(ctx, *source.ID, nil, update)
				if err != nil {
					return nil, errors.Wrap(err, "unable to update data source")
				}
			}
		}

		source = sources[0]
	} else {
		create := dataSource.NewCreate()
		create.ProviderType = pointer.FromString(p.Type())
		create.ProviderName = pointer.FromString(p.Name())
		create.State = pointer.FromString(dataSource.StateDisconnected)

		source, err = p.dataSourceClient.Create(ctx, userID, create)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create data source")
		}
	}

	taskCreate, err := fetch.NewTaskCreate(providerSession.ID, *source.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create task create")
	}

	task, err := p.taskClient.CreateTask(ctx, taskCreate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create task")
	}

	// Update data source to connected after task successfully created
	update := dataSource.NewUpdate()
	update.ProviderSessionID = pointer.FromString(providerSession.ID)
	update.State = pointer.FromString(dataSource.StateConnected)
	if _, err = p.dataSourceClient.Update(ctx, *source.ID, nil, update); err != nil {

		// Attempt to delete task if data source not marked as connected
		if taskErr := p.taskClient.DeleteTask(context.WithoutCancel(ctx), task.ID); taskErr != nil {
			logger.WithError(taskErr).Error("Failure deleting task after failed data source update")
		}

		return nil, errors.Wrap(err, "unable to update data source")
	}

	return nil, nil
}

func (p *Provider) OnDelete(ctx context.Context, userID string, providerSession *auth.ProviderSession) error {
	if userID == "" {
		return errors.New("user id is missing")
	}
	if providerSession == nil {
		return errors.New("provider session is missing")
	}

	logger := log.LoggerFromContext(ctx)

	taskFilter := task.NewTaskFilter()
	taskFilter.Name = pointer.FromString(fetch.TaskName(providerSession.ID))
	tasks, err := p.taskClient.ListTasks(ctx, taskFilter, nil)
	if err != nil {
		logger.WithError(err).Error("unable to list tasks while deleting provider session")
		return nil
	}

	for _, task := range tasks {
		if dataSourceID, ok := task.Data[dexcom.DataKeyDataSourceID].(string); ok && dataSourceID != "" {
			update := dataSource.NewUpdate()
			update.State = pointer.FromString(dataSource.StateDisconnected)
			_, err = p.dataSourceClient.Update(ctx, dataSourceID, nil, update)
			if err != nil {
				logger.WithError(err).WithField(dexcom.DataKeyDataSourceID, dataSourceID).Error("Unable to update data source while deleting provider session")
			}
		}
		if err = p.taskClient.DeleteTask(ctx, task.ID); err != nil {
			logger.WithError(err).WithField("taskId", task.ID).Error("unable to delete task while deleting provider session")
		}
	}
	return nil
}
