package store

import (
	"context"

	"github.com/tidepool-org/platform/task"
	"go.mongodb.org/mongo-driver/mongo"
)

type Store interface {
	NewTaskRepository() TaskRepository
}

type TaskRepository interface {
	task.TaskAccessor

	UpdateFromState(ctx context.Context, tsk *task.Task, state string) (*task.Task, error)
	IteratePending(ctx context.Context) (*mongo.Cursor, error)
}
