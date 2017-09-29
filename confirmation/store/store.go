package store

import (
	"context"

	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewConfirmationsSession() ConfirmationsSession
}

type ConfirmationsSession interface {
	store.Session

	DestroyConfirmationsForUserByID(ctx context.Context, userID string) error
}
