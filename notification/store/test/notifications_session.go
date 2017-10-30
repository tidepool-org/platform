package test

import testStore "github.com/tidepool-org/platform/store/test"

type NotificationsSession struct {
	*testStore.Session
}

func NewNotificationsSession() *NotificationsSession {
	return &NotificationsSession{
		Session: testStore.NewSession(),
	}
}

func (n *NotificationsSession) UnusedOutputsCount() int {
	return n.Session.UnusedOutputsCount()
}
