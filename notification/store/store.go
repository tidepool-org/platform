package store

import "io"

type Store interface {
	NewNotificationsSession() NotificationsSession
}

type NotificationsSession interface {
	io.Closer
}
