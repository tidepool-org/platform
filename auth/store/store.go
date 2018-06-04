package store

import (
	"io"

	"github.com/tidepool-org/platform/auth"
)

type Store interface {
	NewProviderSessionSession() ProviderSessionSession
	NewRestrictedTokenSession() RestrictedTokenSession
}

type ProviderSessionSession interface {
	io.Closer
	auth.ProviderSessionAccessor
}

type RestrictedTokenSession interface {
	io.Closer
	auth.RestrictedTokenAccessor
}
