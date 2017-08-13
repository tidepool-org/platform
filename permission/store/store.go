package store

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewPermissionsSession(logger log.Logger) PermissionsSession
}

type PermissionsSession interface {
	store.Session

	DestroyPermissionsForUserByID(userID string) error
}
