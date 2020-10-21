package store

import (
	"context"
)

type Store interface {
	NewMessageRepository() MessageRepository
}

type MessageRepository interface {
	DeleteMessagesFromUser(ctx context.Context, user *User) error
	DestroyMessagesForUserByID(ctx context.Context, userID string) error
}

// TODO: Temporary until User is restructured

type User struct {
	ID       string
	FullName string
}
