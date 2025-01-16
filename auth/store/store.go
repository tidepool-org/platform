package store

import (
	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/devicetokens"
)

type Store interface {
	NewProviderSessionRepository() ProviderSessionRepository
	NewRestrictedTokenRepository() RestrictedTokenRepository
	NewDeviceTokenRepository() DeviceTokenRepository
	NewAppValidateRepository() appvalidate.Repository
}

type ProviderSessionRepository interface {
	auth.ProviderSessionAccessor
}

type RestrictedTokenRepository interface {
	auth.RestrictedTokenAccessor
}

type DeviceTokenRepository interface {
	devicetokens.Repository
}
