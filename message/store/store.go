package store

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewMessagesSession(logger log.Logger) MessagesSession
}

type MessagesSession interface {
	store.Session

	DeleteMessagesFromUser(user *User) error
	DestroyMessagesForUserByID(userID string) error
}

// TODO: Temporary until User is restructured

type User struct {
	ID       string
	FullName string
}
