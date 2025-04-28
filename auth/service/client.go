package service

import (
	"context"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/page"
)

type Client interface {
	auth.Client

	ListAllProviderSessions(ctx context.Context, filter auth.ProviderSessionFilter, pagination page.Pagination) (auth.ProviderSessions, error)
	DeleteAllProviderSessions(ctx context.Context, filter auth.ProviderSessionFilter) error
}
