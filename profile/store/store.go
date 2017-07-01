package store

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/profile"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewSession(logger log.Logger) Session
}

type Session interface {
	store.Session

	GetProfileByID(profileID string) (*profile.Profile, error)
	DestroyProfileByID(profileID string) error
}
