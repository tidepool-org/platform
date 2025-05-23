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

	BeforeCreate(ctx context.Context, userID string, providerSession *auth.ProviderSessionCreate) error
	OnCreate(ctx context.Context, userID string, providerSession *auth.ProviderSession) error
	OnDelete(ctx context.Context, userID string, providerSession *auth.ProviderSession) error
}
