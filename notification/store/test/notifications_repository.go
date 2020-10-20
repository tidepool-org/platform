package test

type NotificationsRepository struct {
}

func NewNotificationsRepository() *NotificationsRepository {
	return &NotificationsRepository{}
}

func (n *NotificationsRepository) AssertOutputsEmpty() {
}
