package test

import "github.com/tidepool-org/platform/test"

type TasksSession struct {
	*test.Closer
}

func NewTasksSession() *TasksSession {
	return &TasksSession{
		Closer: test.NewCloser(),
	}
}

func (t *TasksSession) AssertOutputsEmpty() {
	t.Closer.AssertOutputsEmpty()
}
