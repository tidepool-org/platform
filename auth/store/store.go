package store

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewProviderSessionSession() ProviderSessionSession
	NewRestrictedTokenSession() RestrictedTokenSession
}

type ProviderSessionSession interface {
	store.Session
	auth.ProviderSessionAccessor
}

type RestrictedTokenSession interface {
	store.Session
	auth.RestrictedTokenAccessor
}
