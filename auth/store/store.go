package store

import (
	"github.com/tidepool-org/platform/auth"
)

type Store interface {
	NewProviderSessionRepository() ProviderSessionRepository
	NewRestrictedTokenRepository() RestrictedTokenRepository
}

type ProviderSessionRepository interface {
	auth.ProviderSessionAccessor
}

type RestrictedTokenRepository interface {
	auth.RestrictedTokenAccessor
}
