package test

import testStore "github.com/tidepool-org/platform/store/test"

type TasksSession struct {
	*testStore.Session
}

func NewTasksSession() *TasksSession {
	return &TasksSession{
		Session: testStore.NewSession(),
	}
}

func (t *TasksSession) UnusedOutputsCount() int {
	return t.Session.UnusedOutputsCount()
}
