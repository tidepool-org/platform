package store

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewSession(logger log.Logger) Session
}

type Session interface {
	store.Session

	DestroyTasksForUserByID(userID string) error
}
