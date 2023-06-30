package store

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/task"
)

type Store interface {
	NewTaskRepository() TaskRepository
}

type TaskRepository interface {
	task.TaskAccessor

	UnstickTasks(ctx context.Context) (int64, error)

	UpdateFromState(ctx context.Context, tsk *task.Task, state string) (*task.Task, error)
	IteratePending(ctx context.Context) (*mongo.Cursor, error)
}
