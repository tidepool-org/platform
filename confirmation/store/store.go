package store

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewConfirmationsSession(logger log.Logger) ConfirmationsSession
}

type ConfirmationsSession interface {
	store.Session

	DestroyConfirmationsForUserByID(userID string) error
}
