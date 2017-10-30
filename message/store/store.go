package store

import (
	"context"

	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewMessagesSession() MessagesSession
}

type MessagesSession interface {
	store.Session

	DeleteMessagesFromUser(ctx context.Context, user *User) error
	DestroyMessagesForUserByID(ctx context.Context, userID string) error
}

// TODO: Temporary until User is restructured

type User struct {
	ID       string
	FullName string
}
