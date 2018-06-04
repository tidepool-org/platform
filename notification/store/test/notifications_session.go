package test

import "github.com/tidepool-org/platform/test"

type NotificationsSession struct {
	*test.Closer
}

func NewNotificationsSession() *NotificationsSession {
	return &NotificationsSession{
		Closer: test.NewCloser(),
	}
}

func (n *NotificationsSession) AssertOutputsEmpty() {
	n.Closer.AssertOutputsEmpty()
}
