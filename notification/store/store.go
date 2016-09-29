package store

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewSession(logger log.Logger) (Session, error)
}

type Session interface {
	store.Session

	DestroyNotificationsForUserByID(userID string) error
}
