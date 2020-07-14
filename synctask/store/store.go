package store

import (
	"context"
)

type Store interface {
	NewSyncTaskRepository() SyncTaskRepository
}

type SyncTaskRepository interface {
	DestroySyncTasksForUserByID(ctx context.Context, userID string) error
}
