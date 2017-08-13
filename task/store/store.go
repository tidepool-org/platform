package store

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewSession(lgr log.Logger) StoreSession
}

type StoreSession interface {
	store.Session
}
