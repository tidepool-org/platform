package store

import (
	"context"
)

type Store interface {
	NewTokenRepository() TokenRepository
}

type TokenRepository interface {
	DestroySessionsForUserByID(ctx context.Context, userID string) error
}
