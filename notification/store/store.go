package store

import (
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewNotificationsSession() NotificationsSession
}

type NotificationsSession interface {
	store.Session
}
