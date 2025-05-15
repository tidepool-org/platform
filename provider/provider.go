package provider

import (
	"context"

	"github.com/tidepool-org/platform/auth"
)

type Factory interface {
	Get(typ string, name string) (Provider, error)
}

type Provider interface {
	Type() string
	Name() string

	OnCreate(ctx context.Context, providerSession *auth.ProviderSession) error
	OnDelete(ctx context.Context, providerSession *auth.ProviderSession) error
}
