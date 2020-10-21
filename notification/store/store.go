package store

type Store interface {
	NewNotificationsRepository() NotificationsRepository
}

type NotificationsRepository interface {
}
