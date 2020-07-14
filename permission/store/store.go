package store

import (
	"context"
)

type Store interface {
	NewPermissionsRepository() PermissionsRepository
}

type PermissionsRepository interface {
	DestroyPermissionsForUserByID(ctx context.Context, userID string) error
}
