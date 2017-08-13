package store

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewAuthsSession(lgr log.Logger) AuthsSession
}

type AuthsSession interface {
	store.Session
}
