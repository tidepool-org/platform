package store

import (
	"context"
	"io"
)

type Store interface {
	NewSyncTaskSession() SyncTaskSession
}

type SyncTaskSession interface {
	io.Closer

	DestroySyncTasksForUserByID(ctx context.Context, userID string) error
}
