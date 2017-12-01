package store

import (
	"github.com/tidepool-org/platform/confirmation"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewConfirmationSession() ConfirmationSession
}

type ConfirmationSession interface {
	store.Session
	confirmation.ConfirmationAccessor
}
