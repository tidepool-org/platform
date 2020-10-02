package events

import (
	"context"
	ev "github.com/tidepool-org/go-common/events"
	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
	dataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type userDeletionEventsHandler struct {
	ev.NoopUserEventsHandler

	ctx       context.Context
	dataStore dataStoreDEPRECATED.Store
	dataSourceStore dataSourceStoreStructured.Store
}

func NewUserDataDeletionHandler(ctx context.Context, dataStore dataStoreDEPRECATED.Store, dataSourceStore dataSourceStoreStructured.Store) ev.EventHandler {
	return ev.NewUserEventsHandler(&userDeletionEventsHandler{
		ctx:       ctx,
		dataStore: dataStore,
		dataSourceStore: dataSourceStore,
	})
}

func (u *userDeletionEventsHandler) HandleDeleteUserEvent(payload ev.DeleteUserEvent) error {
	var errs []error
	logger := log.LoggerFromContext(u.ctx).WithField("userId", payload.UserID)

	logger.Infof("Deleting data for user")
	dataSsn := u.dataStore.NewDataSession()
	defer dataSsn.Close()
	if err := dataSsn.DestroyDataForUserByID(u.ctx, payload.UserID); err != nil {
		errs = append(errs, err)
		logger.WithError(err).Error("unable to delete data for user")
	}

	logger.Infof("Deleting data source for user")
	dataSourceSsn := u.dataSourceStore.NewSession()
	defer dataSourceSsn.Close()
	if _, err := dataSourceSsn.DestroyAll(u.ctx, payload.UserID); err != nil {
		errs = append(errs, err)
		logger.WithError(err).Error("unable to delete data sources for user")
	}

	if len(errs) != 0 {
		return errors.New("Unable to delete device data for user")
	}
	return nil
}
