package store

import (
	"context"
	"io"
)

type Store interface {
	NewMessagesSession() MessagesSession
}

type MessagesSession interface {
	io.Closer

	DeleteMessagesFromUser(ctx context.Context, user *User) error
	DestroyMessagesForUserByID(ctx context.Context, userID string) error
}

// TODO: Temporary until User is restructured

type User struct {
	ID       string
	FullName string
}
