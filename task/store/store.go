package store

import (
	"context"
	"io"

	"github.com/tidepool-org/platform/task"
)

type Store interface {
	NewTaskSession() TaskSession
}

type TaskSession interface {
	io.Closer
	task.TaskAccessor

	UpdateFromState(ctx context.Context, tsk *task.Task, state string) (*task.Task, error)
	IteratePending(ctx context.Context) TaskIterator
}

type TaskIterator interface {
	Next(tsk *task.Task) bool
	Close() error
	Error() error
}
