package store

import (
	"context"

	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/page"
)

type Store interface {
	NewProviderSessionRepository() ProviderSessionRepository
	NewRestrictedTokenRepository() RestrictedTokenRepository
	NewDeviceTokenRepository() DeviceTokenRepository
	NewAppValidateRepository() appvalidate.Repository
}

type ProviderSessionRepository interface {
	auth.ProviderSessionAccessor

	ListAllProviderSessions(ctx context.Context, filter auth.ProviderSessionFilter, pagination page.Pagination) (auth.ProviderSessions, error)
}

type RestrictedTokenRepository interface {
	auth.RestrictedTokenAccessor
}

type DeviceTokenRepository interface {
	devicetokens.Repository
}
