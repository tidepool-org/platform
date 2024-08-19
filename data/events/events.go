package events

import (
	"context"

	ev "github.com/tidepool-org/go-common/events"

	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
	dataStore "github.com/tidepool-org/platform/data/store"
	summaryStore "github.com/tidepool-org/platform/data/summary/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type userDeletionEventsHandler struct {
	ev.NoopUserEventsHandler

	ctx             context.Context
	dataStore       dataStore.Store
	dataSourceStore dataSourceStoreStructured.Store
}

func NewUserDataDeletionHandler(ctx context.Context, dataStore dataStore.Store, dataSourceStore dataSourceStoreStructured.Store) ev.EventHandler {
	return ev.NewUserEventsHandler(&userDeletionEventsHandler{
		ctx:             ctx,
		dataStore:       dataStore,
		dataSourceStore: dataSourceStore,
	})
}

func (u *userDeletionEventsHandler) HandleDeleteUserEvent(payload ev.DeleteUserEvent) error {
	var errs []error
	logger := log.LoggerFromContext(u.ctx).WithField("userId", payload.UserID)

	logger.Infof("Deleting data for user")
	dataRepository := u.dataStore.NewDataRepository()
	if err := dataRepository.DestroyDataForUserByID(u.ctx, payload.UserID); err != nil {
		errs = append(errs, err)
		logger.WithError(err).Error("unable to delete data for user")
	}

	logger.Infof("Deleting data source for user")
	dataSourceRepository := u.dataSourceStore.NewDataSourcesRepository()
	if _, err := dataSourceRepository.DestroyAll(u.ctx, payload.UserID); err != nil {
		errs = append(errs, err)
		logger.WithError(err).Error("unable to delete data sources for user")
	}

	logger.Infof("Deleting summary for user")
	summaryRepository := summaryStore.NewTypeless(u.dataStore.NewSummaryRepository().GetStore())
	if err := summaryRepository.DeleteSummary(u.ctx, payload.UserID); err != nil {
		errs = append(errs, err)
		logger.WithError(err).Error("unable to delete summary for user")
	}

	if len(errs) != 0 {
		return errors.New("Unable to delete device data for user")
	}
	return nil
}
