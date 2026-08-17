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
	ListProviderSessions(ctx context.Context, filter *auth.ProviderSessionFilter, pagination *page.Pagination) (auth.ProviderSessions, error)

	CreateProviderSession(ctx context.Context, create *auth.ProviderSessionCreate) (*auth.ProviderSession, error)
	GetProviderSession(ctx context.Context, id string) (*auth.ProviderSession, error)
	UpdateProviderSession(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error)
	DeleteProviderSession(ctx context.Context, id string) error
}

type RestrictedTokenRepository interface {
	auth.RestrictedTokenAccessor
}

type DeviceTokenRepository interface {
	devicetokens.Repository
}
