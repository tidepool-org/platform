package store

import (
	"context"
	"io"

	"github.com/tidepool-org/platform/user"
)

type Store interface {
	NewUsersSession() UsersSession
}

type UsersSession interface {
	io.Closer

	GetUserByID(ctx context.Context, userID string) (*user.User, error)
	DeleteUser(ctx context.Context, user *user.User) error
	DestroyUserByID(ctx context.Context, userID string) error

	PasswordMatches(user *user.User, password string) bool
}
