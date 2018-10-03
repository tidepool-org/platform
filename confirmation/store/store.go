package store

import (
	"io"

	"github.com/tidepool-org/platform/confirmation"
)

type Store interface {
	NewConfirmationSession() ConfirmationSession
}

type ConfirmationSession interface {
	io.Closer
	confirmation.ConfirmationAccessor
}
