package test

import storeStructuredTest "github.com/tidepool-org/platform/store/structured/test"

type TasksSession struct {
	*storeStructuredTest.Session
}

func NewTasksSession() *TasksSession {
	return &TasksSession{
		Session: storeStructuredTest.NewSession(),
	}
}

func (t *TasksSession) UnusedOutputsCount() int {
	return t.Session.UnusedOutputsCount()
}
