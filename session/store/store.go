package store

import (
	"context"

	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewSessionsSession() SessionsSession
}

type SessionsSession interface {
	store.Session

	DestroySessionsForUserByID(ctx context.Context, userID string) error
}
