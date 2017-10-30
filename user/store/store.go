package store

import (
	"context"

	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/user"
)

type Store interface {
	store.Store

	NewUsersSession() UsersSession
}

type UsersSession interface {
	store.Session

	GetUserByID(ctx context.Context, userID string) (*user.User, error)
	DeleteUser(ctx context.Context, user *user.User) error
	DestroyUserByID(ctx context.Context, userID string) error

	PasswordMatches(user *user.User, password string) bool
}
