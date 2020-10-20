package store

import (
	"github.com/tidepool-org/platform/confirmation"
)

type Store interface {
	NewConfirmationRepository() ConfirmationRepository
}

type ConfirmationRepository interface {
	confirmation.ConfirmationAccessor
}
