package store

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
	"github.com/tidepool-org/platform/user"
)

type Store interface {
	store.Store

	NewSession(logger log.Logger) (Session, error)
}

type Session interface {
	store.Session

	GetUserByID(userID string) (*user.User, error)
	DeleteUser(user *user.User) error
	DestroyUserByID(userID string) error

	PasswordMatches(user *user.User, password string) bool
}
