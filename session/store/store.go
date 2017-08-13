package store

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewSessionsSession(logger log.Logger) SessionsSession
}

type SessionsSession interface {
	store.Session

	DestroySessionsForUserByID(userID string) error
}
