package store

import (
	"context"
	"io"

	"github.com/tidepool-org/platform/profile"
)

type Store interface {
	NewProfilesSession() ProfilesSession
}

type ProfilesSession interface {
	io.Closer

	GetProfileByID(ctx context.Context, profileID string) (*profile.Profile, error)
	DestroyProfileByID(ctx context.Context, profileID string) error
}
