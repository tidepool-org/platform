package store

import (
	"context"
	"io"
)

type Store interface {
	NewPermissionsSession() PermissionsSession
}

type PermissionsSession interface {
	io.Closer

	DestroyPermissionsForUserByID(ctx context.Context, userID string) error
}
