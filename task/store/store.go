package store

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewTasksSession(lgr log.Logger) TasksSession
}

type TasksSession interface {
	store.Session
}
