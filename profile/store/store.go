package store

import (
	"context"

	"github.com/tidepool-org/platform/profile"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewProfilesSession() ProfilesSession
}

type ProfilesSession interface {
	store.Session

	GetProfileByID(ctx context.Context, profileID string) (*profile.Profile, error)
	DestroyProfileByID(ctx context.Context, profileID string) error
}
