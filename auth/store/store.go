package store

import (
	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/auth"
)

type Store interface {
	NewProviderSessionRepository() ProviderSessionRepository
	NewRestrictedTokenRepository() RestrictedTokenRepository
	NewAppValidateRepository() appvalidate.Repository
}

type ProviderSessionRepository interface {
	auth.ProviderSessionAccessor
}

type RestrictedTokenRepository interface {
	auth.RestrictedTokenAccessor
}
