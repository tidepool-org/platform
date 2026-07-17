package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/page"
	storeStructured "github.com/tidepool-org/platform/store/structured"
	"github.com/tidepool-org/platform/task"
)

type Store interface {
	NewTaskRepository() TaskRepository
	WithTypeFilter(typeFilter string) Store
	Terminate(ctx context.Context) error
}

type TaskRepository interface {
	ListTasks(ctx context.Context, filter *task.TaskFilter, pagination *page.Pagination) (task.Tasks, error)
	CreateTask(ctx context.Context, create *task.TaskCreate) (*task.Task, error)
	GetTask(ctx context.Context, id string, condition *storeStructured.Condition) (*task.Task, error)
	UpdateTask(ctx context.Context, id string, condition *storeStructured.Condition, update *task.TaskUpdate) (*task.Task, error)
	DeleteTask(ctx context.Context, id string, condition *storeStructured.Condition) error

	UnstickTasks(ctx context.Context) ([]string, error)

	StartTask(ctx context.Context, id string, revision int, deadline time.Duration) (*task.Task, error)
	StopTask(ctx context.Context, id string, stateLock *string, state string, duration *time.Duration, update *task.TaskUpdate) error

	IteratePending(ctx context.Context) (*mongo.Cursor, error)
}
