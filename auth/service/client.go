package service

import (
	"context"

	"github.com/tidepool-org/platform/auth"
)

type Client interface {
	auth.Client

	DeleteAllProviderSessions(ctx context.Context, filter auth.ProviderSessionFilter) error
}
