package store

import (
	"context"

	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/task"
)

type Store interface {
	store.Store

	NewTaskSession() TaskSession
}

type TaskSession interface {
	store.Session
	task.TaskAccessor

	UpdateFromState(ctx context.Context, tsk *task.Task, state string) (*task.Task, error)
	IteratePending(ctx context.Context) TaskIterator
}

type TaskIterator interface {
	Next(tsk *task.Task) bool
	Close() error
	Error() error
}
