package store

import (
	"context"

	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewSyncTaskSession() SyncTaskSession
}

type SyncTaskSession interface {
	store.Session

	DestroySyncTasksForUserByID(ctx context.Context, userID string) error
}
