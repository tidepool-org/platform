package provider

import "context"

type Factory interface {
	Get(typ string, name string) (Provider, error)
}

type Provider interface {
	Type() string
	Name() string

	OnCreate(ctx context.Context, userID string, providerSessionID string) error
	OnDelete(ctx context.Context, userID string, providerSessionID string) error
}
