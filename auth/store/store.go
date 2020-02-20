package store

import (
	"io"

	"github.com/tidepool-org/platform/auth"
)

type Store interface {
	NewProviderSessionSession() ProviderSessionSession
	NewRestrictedTokenSession() RestrictedTokenSession
	NewDeviceAuthorizationSession() DeviceAuthorizationSession
}

type ProviderSessionSession interface {
	io.Closer
	auth.ProviderSessionAccessor
}

type RestrictedTokenSession interface {
	io.Closer
	auth.RestrictedTokenAccessor
}

type DeviceAuthorizationSession interface {
	io.Closer
	auth.DeviceAuthorizationAccessor
}