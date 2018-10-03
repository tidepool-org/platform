package store

import (
	"context"
	"io"
)

type Store interface {
	NewSessionsSession() SessionsSession
}

type SessionsSession interface {
	io.Closer

	DestroySessionsForUserByID(ctx context.Context, userID string) error
}
