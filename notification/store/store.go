package store

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewNotificationsSession(lgr log.Logger) NotificationsSession
}

type NotificationsSession interface {
	store.Session
}
