package store

import (
	"context"

	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewPermissionsSession() PermissionsSession
}

type PermissionsSession interface {
	store.Session

	DestroyPermissionsForUserByID(ctx context.Context, userID string) error
}
