package store

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewSyncTasksSession(logger log.Logger) SyncTasksSession
}

type SyncTasksSession interface {
	store.Session

	DestroySyncTasksForUserByID(userID string) error
}
